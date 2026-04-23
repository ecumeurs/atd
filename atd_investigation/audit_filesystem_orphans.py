import os
import re

atoms_dir = "docs"
code_dirs = ["upsilonapi", "upsiloncli", "upsilonbattle", "battleui", "scripts"]

ids = []
atom_data = {}

# 1. Get all IDs from docs/*.atom.md
for filename in os.listdir(atoms_dir):
    if filename.endswith(".atom.md"):
        path = os.path.join(atoms_dir, filename)
        with open(path, "r") as f:
            content = f.read()
            # Extract frontmatter
            match = re.search(r"---(.*?)---", content, re.DOTALL)
            if match:
                fm = match.group(1)
                id_match = re.search(r"id:\s*(.*)", fm)
                type_match = re.search(r"type:\s*(.*)", fm)
                layer_match = re.search(r"layer:\s*(.*)", fm)
                status_match = re.search(r"status:\s*(.*)", fm)
                
                aid = id_match.group(1).strip() if id_match else None
                atype = type_match.group(1).strip() if type_match else "UNKNOWN"
                alayer = layer_match.group(1).strip() if layer_match else "UNKNOWN"
                astatus = status_match.group(1).strip() if status_match else "UNKNOWN"
                
                if aid:
                    ids.append(aid)
                    atom_data[aid] = {"type": atype, "layer": alayer, "status": astatus, "file": path}

# 2. Search for tags in code
implementations = {}
for aid in ids:
    implementations[aid] = False

for d in code_dirs:
    if not os.path.exists(d): continue
    for root, _, files in os.walk(d):
        for file in files:
            if file.endswith((".go", ".js", ".vue", ".php", ".py")):
                path = os.path.join(root, file)
                try:
                    with open(path, "r") as f:
                        c = f.read()
                        for aid in ids:
                            if f"[[{aid}]]" in c:
                                implementations[aid] = True
                except:
                    pass

# 3. Report
orphans = [aid for aid in ids if not implementations[aid] and atom_data[aid]["status"] == "STABLE"]
unknowns = [aid for aid in ids if atom_data[aid]["type"] == "UNKNOWN" or atom_data[aid]["layer"] == "UNKNOWN"]

print(f"Total Atoms Checked: {len(ids)}")
print(f"Total STABLE Orphans: {len(orphans)}")
print(f"Total UNKNOWN Metadata: {len(unknowns)}")

# Group orphans by type
type_counts = {}
for aid in orphans:
    t = atom_data[aid]["type"]
    type_counts[t] = type_counts.get(t, 0) + 1

print("\nOrphan Breakdown by Type:")
for t, count in type_counts.items():
    print(f"{t}: {count}")

print("\nOrphan List:")
for aid in orphans:
    print(f"- {aid} ({atom_data[aid]['type']}/{atom_data[aid]['layer']})")

print("\nUNKNOWN List:")
for aid in unknowns:
    print(f"- {aid} (Type: {atom_data[aid]['type']}, Layer: {atom_data[aid]['layer']})")
