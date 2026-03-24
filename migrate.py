#!/usr/bin/env python3
import os
import re
import argparse
import sys

# Priority mapping: Enum -> Numeric
PRIORITY_MAP = {
    "CORE": "5",
    "SECONDARY": "3",
    "EXPERIMENTAL": "2",
    "FLAVOR": "1"
}

# Layer mapping: Type -> Layer
TYPE_TO_LAYER = {
    "REQUIREMENT": "CUSTOMER",
    "SPECIFICATION": "CUSTOMER",
    "USECASE": "CUSTOMER",
    "USER_STORY": "CUSTOMER",
    "DOMAIN": "CUSTOMER",
    "USAGE": "CUSTOMER",
    "MODULE": "ARCHITECTURE",
    "SERVICE": "ARCHITECTURE",
    "ENTITY": "ARCHITECTURE",
    "API": "ARCHITECTURE",
    "UI": "ARCHITECTURE",
    "RULE": "ARCHITECTURE",  # Defaulting to ARCHITECTURE for rules; can be manually refined
    "MECHANIC": "IMPLEMENTATION",
    "DATA": "IMPLEMENTATION",
    "BUILD": "IMPLEMENTATION"
}

def migrate_file(filepath, dry_run=False):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    # Split frontmatter
    parts = re.split(r'^---\s*$', content, flags=re.MULTILINE)
    if len(parts) < 3:
        print(f"Skipping {filepath}: No valid frontmatter found.")
        return False

    frontmatter = parts[1]
    body = "---".join(parts[2:])

    # Check for existing layer field
    if re.search(r'^layer:', frontmatter, re.MULTILINE):
        # Even if layer exists, ensure priority is numeric
        pass
    else:
        # Determine layer based on type
        type_match = re.search(r'^type:\s*(\w+)', frontmatter, re.MULTILINE)
        if type_match:
            atom_type = type_match.group(1).upper()
            layer = TYPE_TO_LAYER.get(atom_type, "IMPLEMENTATION")
            
            # Insert layer after type
            frontmatter = re.sub(
                r'^(type:\s*.*)$',
                f'\\1\nlayer: {layer}',
                frontmatter,
                flags=re.MULTILINE
            )
        else:
            print(f"Warning: No type found in {filepath}. Skipping layer assignment.")

    # Convert priority
    priority_match = re.search(r'^priority:\s*(\w+)', frontmatter, re.MULTILINE)
    if priority_match:
        old_prio = priority_match.group(1).upper()
        if old_prio in PRIORITY_MAP:
            new_prio = PRIORITY_MAP[old_prio]
            frontmatter = re.sub(
                f'^priority:\\s*{old_prio}',
                f'priority: {new_prio}',
                frontmatter,
                flags=re.MULTILINE
            )

    new_content = f"---{frontmatter}---{body}"

    if content == new_content:
        return False

    if dry_run:
        print(f"[DRY-RUN] Would update {filepath}")
    else:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(new_content)
        print(f"Updated {filepath}")
    
    return True

def main():
    parser = argparse.ArgumentParser(description="Migrate ATD atoms to the new layer/priority standard.")
    parser.add_argument("path", help="Directory containing .atom.md files")
    parser.add_argument("--dry-run", action="store_true", help="Show what would be changed without writing")
    args = parser.parse_args()

    if not os.path.isdir(args.path):
        print(f"Error: {args.path} is not a directory.")
        sys.exit(1)

    count = 0
    for root, _, files in os.walk(args.path):
        for file in files:
            if file.endswith(".atom.md"):
                filepath = os.path.join(root, file)
                if migrate_file(filepath, args.dry_run):
                    count += 1

    print(f"\nMigration complete. Total files processed: {count}")

if __name__ == "__main__":
    main()
