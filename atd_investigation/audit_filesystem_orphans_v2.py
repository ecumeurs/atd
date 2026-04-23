import os
import re

atoms_dir = "docs"
code_dirs = ["upsilonapi", "upsiloncli", "upsilonbattle", "battleui", "scripts"]

ids = []
atom_data = {}

# 1. Get all IDs and Metadata from docs/*.atom.md
for filename in os.listdir(atoms_dir):
    if filename.endswith(".atom.md"):
        path = os.path.join(atoms_dir, filename)
        with open(path, "r") as f:
            content = f.read()
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

# 3. Apply Rules
# IMPL must have links. ARCH should have links.
must_have_links = [aid for aid in ids if atom_data[aid]["layer"] == "IMPLEMENTATION" and atom_data[aid]["status"] == "STABLE" and not implementations[aid]]
should_have_links = [aid for aid in ids if atom_data[aid]["layer"] == "ARCHITECTURE" and atom_data[aid]["status"] == "STABLE" and not implementations[aid]]

print(f"Total Atoms: {len(ids)}")
print(f"Total STABLE Orphans: {len([aid for aid in ids if not implementations[aid] and atom_data[aid]['status'] == 'STABLE'])}")
print("\n--- MANDATORY (IMPLEMENTATION) Orphans ---")
for aid in must_have_links:
    print(f"- {aid} ({atom_data[aid]['type']})")

print("\n--- HIGH-PROBABILITY (ARCHITECTURE) Orphans ---")
for aid in should_have_links:
    print(f"- {aid} ({atom_data[aid]['type']})")

