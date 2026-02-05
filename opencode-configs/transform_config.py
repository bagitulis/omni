#!/usr/bin/env python3
"""Transform oh-my-opencode config for plugin mode.

Replaces google/claude-* and google/gemini-* with google/antigravity-* prefixes.
"""

import sys

def transform_for_plugin(content: str) -> str:
    """Transform model names for plugin mode."""
    content = content.replace('google/claude-', 'google/antigravity-claude-')
    content = content.replace('google/gemini-', 'google/antigravity-gemini-')
    return content

def main():
    if len(sys.argv) != 3:
        print("Usage: transform_config.py <source_file> <dest_file>")
        sys.exit(1)
    
    src_file = sys.argv[1]
    dst_file = sys.argv[2]
    
    try:
        with open(src_file, 'r', encoding='utf-8') as f:
            content = f.read()
        
        transformed = transform_for_plugin(content)
        
        with open(dst_file, 'w', encoding='utf-8') as f:
            f.write(transformed)
        
        print(f"Transformed {src_file} -> {dst_file}")
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

if __name__ == '__main__':
    main()
