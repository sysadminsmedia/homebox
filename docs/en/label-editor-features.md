# Label Editor Features

This document tracks the features of the label template editor.

## Implemented Features (v0.1)

### Label Template System
- Create, edit, duplicate, and delete label templates
- Support for custom dimensions or preset label sizes
- Presets for common label formats:
  - **Avery** sheet labels (5160, 5163, 5167, etc.)
  - **Brother** continuous and die-cut labels (DK series)
  - **Dymo** labels (various sizes)
- Shared templates within groups
- Default template for quick printing

### Canvas Editor
- Visual WYSIWYG label design
- Add text elements with data field placeholders
- Add barcodes (QR Code, Code128, DataMatrix, etc.)
- Add shapes (rectangles, lines)
- Drag and resize elements
- Rotation support

### Text Formatting
- Font size control
- Bold text toggle
- Text alignment (left, center, right)
- Text color picker
- Auto-fit text (scales font to fit within bounds)
- Insert data field placeholders:
  - Item fields: name, description, asset ID, serial number, etc.
  - Location fields: name, path, item count
  - Common fields: current date, current time

### Layer Management
- Layers panel showing all canvas objects
- Reorder objects (bring forward, send backward, to front, to back)
- Click layer items to select objects on canvas
- Toggle layer visibility

### Grid and Snap
- Optional grid overlay on canvas
- Snap objects to grid when moving/resizing
- Configurable grid size

### Ruler
- Rulers along top and left edges of canvas
- Display measurements in millimeters or inches
- Zoom-aware ruler scaling

### Zoom Controls
- Zoom in/out buttons
- Fit to view
- Reset to 100%
- Zoom range: 25% to 400%

### Live Preview
- Real-time preview with actual item data
- Preview updates automatically when canvas changes (debounced)
- No need to save before previewing

### Printing
- Download as PNG or PDF
- Multi-page PDF support for batch printing
- Sheet layout support (Avery-style label sheets)
- Cut guides option for manual cutting
- Direct printing to network printers (IPP/CUPS)
- Direct printing to Brother label printers (raster protocol)

### Printer Management
- Add and configure network printers
- Support for IPP/CUPS printers
- Support for Brother QL-series label printers
- Printer status checking
- Test print functionality
- Default printer selection

## Future Considerations

### Object Alignment Tools
- Align selected objects (left, center, right, top, middle, bottom)
- Distribute objects evenly
- Match width/height of selected objects

### Undo/Redo
- Track canvas state history
- Keyboard shortcuts (Ctrl+Z, Ctrl+Y)
- Visual undo history panel

### Templates Library
- Pre-built template designs
- Import/export templates
- Template categories and search

### Advanced Text Features
- Multi-line text with line breaks
- Text along path (curved text)
- Character spacing control
- Text outline/shadow effects

### Smart Guides
- Snap to other objects' edges
- Alignment guides when moving objects
