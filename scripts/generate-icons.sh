#!/bin/bash

# Generate PWA icons for CasGists
# This script creates placeholder icons - replace with actual SVG to PNG conversion

ICON_DIR="web/static/icons"
mkdir -p "$ICON_DIR"

# Base64 encoded 1x1 PNG for placeholder icons
BASE64_PNG="iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChAI9jU8w2wAAAABJRU5ErkJggg=="

# Function to create a placeholder icon of given size
create_icon() {
    local size=$1
    local filename="$ICON_DIR/icon-${size}x${size}.png"
    
    # Create a colored square icon using ImageMagick (if available) or placeholder
    if command -v convert >/dev/null 2>&1; then
        convert -size ${size}x${size} xc:"#6366f1" -fill white -gravity center -pointsize $((size/4)) -annotate +0+0 "CG" "$filename"
    else
        # Fallback: create a minimal PNG file
        echo "$BASE64_PNG" | base64 -d > "$filename"
    fi
    
    echo "Created $filename"
}

# Generate all required icon sizes
SIZES=(72 96 128 144 152 192 384 512)

echo "Generating PWA icons for CasGists..."

for size in "${SIZES[@]}"; do
    create_icon "$size"
done

# Create additional icons
if command -v convert >/dev/null 2>&1; then
    # Badge icon (smaller, optimized for notifications)
    convert -size 72x72 xc:"#6366f1" -fill white -gravity center -pointsize 32 -annotate +0+0 "C" "$ICON_DIR/badge-72x72.png"
    
    # Action icons
    convert -size 96x96 xc:"#10b981" -fill white -gravity center -pointsize 48 -annotate +0+0 "+" "$ICON_DIR/new-gist-icon.png"
    convert -size 96x96 xc:"#3b82f6" -fill white -gravity center -pointsize 48 -annotate +0+0 "G" "$ICON_DIR/my-gists-icon.png"
    convert -size 96x96 xc:"#f59e0b" -fill white -gravity center -pointsize 48 -annotate +0+0 "?" "$ICON_DIR/search-icon.png"
    convert -size 96x96 xc:"#ef4444" -fill white -gravity center -pointsize 48 -annotate +0+0 "A" "$ICON_DIR/admin-icon.png"
    convert -size 96x96 xc:"#8b5cf6" -fill white -gravity center -pointsize 48 -annotate +0+0 "V" "$ICON_DIR/view-icon.png"
    convert -size 96x96 xc:"#6b7280" -fill white -gravity center -pointsize 48 -annotate +0+0 "X" "$ICON_DIR/dismiss-icon.png"
    
    # Screenshots (placeholder)
    convert -size 1280x720 xc:"#1e1e2e" -fill "#a6e3a1" -gravity center -pointsize 72 -annotate +0+0 "CasGists Dashboard" "$ICON_DIR/screenshot-wide.png"
    convert -size 640x1136 xc:"#1e1e2e" -fill "#a6e3a1" -gravity center -pointsize 48 -annotate +0+0 "CasGists\nMobile" "$ICON_DIR/screenshot-mobile.png"
else
    echo "ImageMagick not found. Created minimal placeholder icons."
    echo "For production, replace these with proper SVG-generated icons."
    
    # Create placeholder files for missing icons
    for icon in badge-72x72 new-gist-icon my-gists-icon search-icon admin-icon view-icon dismiss-icon; do
        echo "$BASE64_PNG" | base64 -d > "$ICON_DIR/${icon}.png"
    done
    
    # Create larger placeholder screenshots
    dd if=/dev/zero bs=1024 count=10 2>/dev/null | base64 -w 0 | head -c 1000 | base64 -d > "$ICON_DIR/screenshot-wide.png" 2>/dev/null || echo "$BASE64_PNG" | base64 -d > "$ICON_DIR/screenshot-wide.png"
    dd if=/dev/zero bs=1024 count=5 2>/dev/null | base64 -w 0 | head -c 500 | base64 -d > "$ICON_DIR/screenshot-mobile.png" 2>/dev/null || echo "$BASE64_PNG" | base64 -d > "$ICON_DIR/screenshot-mobile.png"
fi

echo "PWA icon generation complete!"
echo ""
echo "Note: These are placeholder icons. For production:"
echo "1. Create a proper SVG logo for CasGists"
echo "2. Use a tool like PWA Asset Generator to create optimized icons"
echo "3. Replace placeholder screenshots with actual app screenshots"