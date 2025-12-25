# Label Templates

Label templates allow you to design custom labels for your items and locations. You can create professional-looking labels with barcodes, text, and data fields that automatically populate with your item information.

## Getting Started

### Accessing Label Templates

Navigate to **Label Templates** from the sidebar menu. Here you can:
- View all your label templates
- Create new templates
- Edit, duplicate, or delete existing templates
- Set a default template for quick printing

### Creating Your First Template

1. Click **Create Template** button
2. Enter a name for your template (e.g., "Asset Label")
3. Choose a label size:
   - Select a **preset** for common label formats (Avery, Brother, Dymo)
   - Or use **Custom Dimensions** for any size
4. Click **Create** to open the label editor

::: tip
If you have a Brother or Dymo label printer, select the matching preset for your label stock. This ensures proper sizing and alignment.
:::

## Label Editor

The label editor provides a visual canvas where you design your label layout.

### Adding Elements

Use the toolbar at the top to add elements:

| Element | Description |
|---------|-------------|
| **Text** | Static text or data field placeholders |
| **Barcode** | QR codes, Code128, DataMatrix, and more |
| **Shape** | Rectangles and lines for borders/separators |

### Working with Text

1. Click **Add Text** to create a text element
2. Select the text element on the canvas
3. Use the **Text Toolbar** (appears above canvas) to:
   - Change font size
   - Toggle bold
   - Set alignment (left, center, right)
   - Pick text color
   - Enable auto-fit (text scales to fit bounds)

### Inserting Data Fields

Data fields are placeholders that get replaced with actual item data when printing.

1. Select a text element
2. Click **Insert Data Field** in the text toolbar
3. Choose from available fields:

**Item Fields:**
- Item Name, Description, Asset ID
- Serial Number, Model Number, Manufacturer
- Location, Location Path
- Labels, Quantity, Notes
- Item URL (for QR codes)

**Location Fields:**
- Location Name, Description, Full Path
- Item Count, Location URL

**Common Fields:**
- Current Date, Current Time

::: info
Data fields appear as `{{field_name}}` in the editor. They're replaced with real data when you preview or print.
:::

### Adding Barcodes

1. Click **Add Barcode** in the toolbar
2. Select the barcode format:
   - **QR Code** - Best for URLs and longer data
   - **Code128** - Common 1D barcode for asset IDs
   - **DataMatrix** - Compact 2D barcode
   - **EAN13/UPC** - Product barcodes
3. Choose the data source (Item URL, Asset ID, Serial Number, etc.)

### Layer Management

The **Layers Panel** (right sidebar) shows all elements on your canvas:

- Click a layer to select that element
- Drag layers to reorder (or use the arrow buttons)
- Click the eye icon to hide/show elements
- Use "Bring to Front" / "Send to Back" for precise ordering

::: tip
If a shape is blocking selection of text underneath, use the Layers panel to select the text, or send the shape to the back.
:::

### Canvas Tools

**Zoom Controls** (below canvas):
- Use +/- buttons to zoom in/out (25% to 400%)
- Click "Fit" to fit the label to the view
- Click "100%" to reset to actual size

**Grid & Snap** (Settings gear icon):
- Enable grid overlay for alignment guides
- Enable snap-to-grid for precise positioning
- Adjust grid size (default: 10px)

**Ruler** (Settings gear icon):
- Toggle ruler visibility
- Switch between metric (mm) and imperial (inches)

### Live Preview

The **Preview Panel** (below canvas) shows how your label will look with real data:

1. Search for and select an item
2. The preview updates automatically as you edit
3. No need to save before previewing

## Printing Labels

### From the Editor

1. Click **Save** to save your template
2. Navigate to an item or location page
3. Click the **Print Label** button
4. Select your template and print options

### Batch Printing

For printing multiple labels at once:

1. Go to **Label Templates** > **Batch Print**
2. Select a template
3. Search and select items to include
4. Choose output format:
   - **PNG** - Single image (or sheet for Avery-style)
   - **PDF** - Multi-page document
5. Adjust quantities per item if needed
6. Click **Download** or **Print**

### Quick Print from Items

On any item page:
1. Click the **Print** button in the item actions
2. If you have a default template set, it prints immediately
3. Otherwise, select a template from the dialog

### Direct Printing to Label Makers

If you've configured a printer (see [Printer Setup](#printer-setup)), you can print directly:

1. Open the Print Dialog
2. Select your printer from the dropdown
3. Click **Print to Label Maker**

## Printer Setup

### Adding a Printer

1. Navigate to **Printers** from the sidebar
2. Click **Add Printer**
3. Enter printer details:
   - **Name** - Friendly name (e.g., "Office Label Printer")
   - **Type** - IPP/CUPS or Brother Raster
   - **Address** - Network address (e.g., `192.168.1.100` or `printer.local`)

### Supported Printers

**IPP/CUPS Printers:**
- Most network printers
- Printers shared via CUPS on Linux/macOS
- AirPrint-compatible printers

**Brother Label Printers:**
- QL-800, QL-810W, QL-820NWB
- QL-1100, QL-1110NWB
- Other QL-series with network connectivity

::: warning
Brother printers must be connected via network (not USB) for direct printing from Homebox.
:::

### Testing Your Printer

1. Go to **Printers**
2. Click the menu on your printer card
3. Select **Test Print**
4. A test label will be sent to verify connectivity

### Setting a Default Printer

1. Go to **Printers**
2. Click the menu on your preferred printer
3. Select **Set as Default**

The default printer will be pre-selected in print dialogs.

## Label Presets

### Sheet Labels (Avery-style)

Sheet label presets include layout information for proper alignment:

| Preset | Size | Layout |
|--------|------|--------|
| Avery 5160 | 1" x 2-5/8" | 3 × 10 per sheet |
| Avery 5163 | 2" x 4" | 2 × 5 per sheet |
| Avery 5167 | 0.5" x 1.75" | 4 × 20 per sheet |

When printing to PDF, labels are automatically arranged in the correct grid layout.

### Brother Labels

| Preset | Size | Type |
|--------|------|------|
| DK-11201 | 29mm × 90mm | Die-cut address |
| DK-11204 | 17mm × 54mm | Die-cut multi-purpose |
| DK-22205 | 62mm wide | Continuous roll |
| DK-22251 | 62mm wide | Red/Black continuous |

::: tip
For Brother printers, select the matching DK label type to ensure proper print dimensions.
:::

### Dymo Labels

Common Dymo LabelWriter sizes are available as presets.

## Tips & Best Practices

### Design Tips

1. **Keep it simple** - Labels are small, don't overcrowd
2. **Use auto-fit** - For variable-length data like item names
3. **Test with real data** - Use the preview to check different items
4. **Consider barcode size** - QR codes need adequate size to scan reliably

### Barcode Guidelines

- **QR Codes**: Minimum 15mm × 15mm for reliable scanning
- **Code128**: Works well for short alphanumeric data
- **Quiet zone**: Leave small margins around barcodes

### Template Organization

- Create templates for different purposes (asset tags, shelf labels, etc.)
- Use descriptive names
- Share templates with your group for consistency
- Set frequently used template as default

## Troubleshooting

### Labels Not Aligning

- Verify you selected the correct preset for your label stock
- Check that your printer margins match the label sheet
- For sheet labels, ensure the PDF page size matches (Letter vs A4)

### Barcode Won't Scan

- Increase barcode size
- Ensure adequate contrast (black on white)
- Check that data doesn't exceed format limits
- Print at highest quality setting

### Printer Connection Issues

- Verify printer IP address is correct
- Ensure printer is on the same network
- Check firewall isn't blocking port 631 (IPP) or 9100 (raw)
- Try the printer's hostname instead of IP

### Preview Not Updating

- Check that an item is selected in the preview panel
- Save the template and refresh the page
- Clear browser cache if issues persist
