# Flat Phase 6 - Directory Metadata & Newline Handling

## Goal

Continue implementing the directory metadata support and newline handling for the flat file format (.fmdx).

## Accomplished

**Completed:**
- ✅ Added `EndWithNewline bool` field to Metadata struct
- ✅ Added `--just-agents` flag support in unflatten
- ✅ Implemented directory metadata collection functions (ReadFlatdir, WriteFlatdir, WriteAgents)
- ✅ Flatten ignores .agents.yaml files (added to automatic ignore patterns)
- ✅ Fixed parser to preserve exact content (readContentBlock)
- ✅ Set `EndWithNewline` based on original file content during flatten
- ✅ Added trailing newline to text files that don't have one
- ✅ Remove trailing newlines during unflatten when `EndWithNewline: false`
- ✅ Fixed parser to skip blank lines when reading content

**Verified:**
- ✅ Round-trip test: flatten → unflatten produces identical files
  - file1.txt (no newline): preserved correctly
  - file2.txt (with newline): preserved correctly
  - binary.bin (with newline): preserved correctly

**Known Issues:**
- ⚠️ Directory entries not being written during flatten (need to implement WriteDirectoryEntry calls)
- ⚠️ AGENTS.yaml creation during unflatten needs directory entries in .fmdx
