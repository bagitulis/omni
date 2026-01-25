#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads - Markdown to HTML Converter
========================================
Convert Markdown content to HTML elements.
"""

import re


def md_to_html(md_content: str) -> str:
    """Enhanced Markdown to HTML conversion."""
    lines = md_content.split('\n')
    html_lines = []
    in_table = False
    in_list = False
    in_code = False
    list_type = None
    
    for line in lines:
        if line.strip() == '```':
            if in_code:
                html_lines.append("</pre>")
                in_code = False
            else:
                html_lines.append("<pre>")
                in_code = True
            continue
        
        if in_code:
            html_lines.append(line)
            continue
        
        if line.startswith('# '):
            html_lines.append(f"<h1>{line[2:]}</h1>")
        elif line.startswith('## '):
            html_lines.append(f"<h2>{line[3:]}</h2>")
        elif line.startswith('### '):
            html_lines.append(f"<h3>{line[4:]}</h3>")
        elif line.startswith('> '):
            html_lines.append(f"<blockquote>{line[2:]}</blockquote>")
        elif line.strip() == '---':
            html_lines.append("<hr>")
        elif line.startswith('|'):
            _process_table_line(line, html_lines, in_table)
            in_table = True
        else:
            if in_table:
                html_lines.append("</tbody></table>")
                in_table = False
            
            in_list, list_type = _process_text_line(
                line, html_lines, in_list, list_type
            )
    
    # Close any open elements
    if in_table:
        html_lines.append("</tbody></table>")
    if in_list:
        html_lines.append(f"</{list_type}>")
    if in_code:
        html_lines.append("</pre>")
    
    return "\n".join(html_lines)


def _process_table_line(line: str, html_lines: list, in_table: bool):
    """Process a table line and add to html_lines."""
    if not in_table:
        html_lines.append("<table>")
    
    if '---' in line:
        return
    
    cells = [c.strip() for c in line.split('|')[1:-1]]
    if len(html_lines) > 0 and html_lines[-1] == "<table>":
        html_lines.append(
            "<thead><tr>" + 
            "".join(f"<th>{c}</th>" for c in cells) + 
            "</tr></thead><tbody>"
        )
    else:
        html_lines.append(
            "<tr>" + 
            "".join(f"<td>{c}</td>" for c in cells) + 
            "</tr>"
        )


def _process_text_line(line: str, html_lines: list, 
                       in_list: bool, list_type: str) -> tuple:
    """Process a text line (not table/code) and return updated list state."""
    if line.startswith('- '):
        if not in_list or list_type != 'ul':
            if in_list:
                html_lines.append(f"</{list_type}>")
            html_lines.append("<ul>")
            in_list = True
            list_type = 'ul'
        html_lines.append(f"<li>{line[2:]}</li>")
    elif re.match(r'^\d+\. ', line):
        if not in_list or list_type != 'ol':
            if in_list:
                html_lines.append(f"</{list_type}>")
            html_lines.append("<ol>")
            in_list = True
            list_type = 'ol'
        html_lines.append(f"<li>{re.sub(r'^\\d+\\. ', '', line)}</li>")
    else:
        if in_list:
            html_lines.append(f"</{list_type}>")
            in_list = False
            list_type = None
        
        # Apply inline formatting
        line = re.sub(r'\*\*(.+?)\*\*', r'<strong>\1</strong>', line)
        line = re.sub(r'\*(.+?)\*', r'<em>\1</em>', line)
        line = re.sub(r'`(.+?)`', r'<code>\1</code>', line)
        
        if line.strip():
            html_lines.append(f"<p>{line}</p>")
    
    return in_list, list_type
