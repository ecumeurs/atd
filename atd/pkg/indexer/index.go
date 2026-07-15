package indexer

import (
	"atd-tools/config"
	atdstore "atd-tools/pkg/store"
	"atd-tools/pkg/ollama"
	"atd-tools/pkg/exploration"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

type IndexReport struct {
	Text string
}

type chunkJob struct {
	path      string
	chunkText string
	modTime   int64
}

func Index(targetDir, dbPath, mode string) (*IndexReport, error) {
	var output strings.Builder

	res, err := ollama.ResolveProvider("embed")
	if err != nil {
		return nil, fmt.Errorf("error resolving embedding provider: %v", err)
	}
	if res.IsIDE {
		return nil, fmt.Errorf("indexing requires a local or remote Ollama provider with embedding model (e.g. nomic-embed-text) installed. No embedding model was found")
	}

	os.MkdirAll(filepath.Dir(dbPath), 0755)

	store, err := atdstore.NewStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open store: %v", err)
	}
	defer store.Close()

	entries, err := store.ListAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list existing entries: %v", err)
	}

	fileMods := make(map[string]int64)
	seenPaths := make(map[string]bool)
	for _, entry := range entries {
		fileMods[entry.AtomPath] = entry.Mtime
	}

	output.WriteString(fmt.Sprintf("Crawling %s (Using %s discovery method)...\n", targetDir, config.GetDiscoveryMethod()))

	var filesToIndex []string
	var codePaths []string

	if len(config.ActiveConfig.CodePaths) > 0 {
		for _, path := range config.ActiveConfig.CodePaths {
			absPath := filepath.Join(targetDir, path)
			codePaths = append(codePaths, absPath)
		}
	} else {
		codePaths = []string{targetDir}
	}

	explorer := exploration.NewExplorer(targetDir, config.DocsDir())
	filesToIndex, err = explorer.ListFiles()
	if err != nil {
		return nil, fmt.Errorf("error discovering files: %v", err)
	}

	jobs := make(chan chunkJob, 100)
	var wg sync.WaitGroup
	var indexedCount int32
	numWorkers := 10

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				emb, err := ollama.QueryEmbed(job.chunkText)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to embed chunk from %s (Error: %v)\n", job.path, err)
					continue
				}

				chunkID := generateChunkID(job.path, job.chunkText)
				err = store.PutEmbedding(atdstore.IndexEntry{
					ChunkID:   chunkID,
					Embedding: emb,
					Mtime:     job.modTime,
					Content:   job.chunkText,
					AtomID:    "",
					AtomPath:  job.path,
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to insert into DB: %v\n", err)
				} else {
					atomic.AddInt32(&indexedCount, 1)
				}
			}
		}()
	}

	var skipCount int
	var fileCount int
	for _, relPath := range filesToIndex {
		if relPath == "" {
			continue
		}

		ext := filepath.Ext(relPath)
		isAtom := strings.HasSuffix(relPath, ".atom.md")

		shouldIndex := false

		switch mode {
		case "code":
			shouldIndex = config.ActiveConfig.SupportedExtensions[ext]
		case "docs":
			shouldIndex = isAtom
		case "all":
			shouldIndex = config.ActiveConfig.SupportedExtensions[ext] || isAtom
		default:
			close(jobs)
			wg.Wait()
			return nil, fmt.Errorf("invalid mode: %s (must be code|docs|all)", mode)
		}

		if !shouldIndex {
			continue
		}

		path := filepath.Join(targetDir, relPath)
		seenPaths[path] = true
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}

		modTime := info.ModTime().Unix()
		if lastMod, exists := fileMods[path]; exists && lastMod == modTime {
			skipCount++
			continue
		}

		store.DeleteByAtomPath(path)

		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var chunks []string
		if isAtom {
			raw := strings.Split(string(content), "## ")
			for _, c := range raw {
				c = strings.TrimSpace(c)
				if c != "" {
					subChunks := splitChunkRecursive("File: "+relPath+"\n## "+c, 8000)
					chunks = append(chunks, subChunks...)
				}
			}
		} else {
			raw := strings.Split(string(content), "\n\n")
			for _, c := range raw {
				c = strings.TrimSpace(c)
				if len(c) > 20 {
					subChunks := splitChunkRecursive("File: "+relPath+"\n"+c, 8000)
					chunks = append(chunks, subChunks...)
				}
			}
		}

		fileCount++
		for _, chunkText := range chunks {
			jobs <- chunkJob{path: path, chunkText: chunkText, modTime: modTime}
		}
	}

	close(jobs)
	wg.Wait()

	output.WriteString(fmt.Sprintf("Indexed %d chunks across %d files (%d skipped unchanged)\n", indexedCount, fileCount, skipCount))
	config.Log("atd-index", fmt.Sprintf("Indexed %d chunks", indexedCount))
	return &IndexReport{Text: output.String()}, nil
}

func splitChunkRecursive(text string, maxLength int) []string {
	if len(text) <= maxLength {
		return []string{text}
	}

	lines := strings.Split(text, "\n")
	if len(lines) > 1 {
		var chunks []string
		current := ""
		for _, line := range lines {
			if len(current)+len(line)+1 > maxLength {
				if current != "" {
					chunks = append(chunks, current)
				}
				if len(line) > maxLength {
					hardChunks := hardSplit(line, maxLength)
					chunks = append(chunks, hardChunks[:len(hardChunks)-1]...)
					current = hardChunks[len(hardChunks)-1]
				} else {
					current = line
				}
			} else {
				if current == "" {
					current = line
				} else {
					current += "\n" + line
				}
			}
		}
		if current != "" {
			chunks = append(chunks, current)
		}
		return chunks
	}

	return hardSplit(text, maxLength)
}

func hardSplit(text string, maxLength int) []string {
	var chunks []string
	for i := 0; i < len(text); i += maxLength {
		end := i + maxLength
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
	}
	return chunks
}

func generateChunkID(path, content string) string {
	h := sha1.New()
	h.Write([]byte(path))
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))[:16]
}