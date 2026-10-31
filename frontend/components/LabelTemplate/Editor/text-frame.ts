import { Textbox } from "fabric";

/** Custom data carried by a text element; the server renderer reads the same keys. */
export interface TextBoxData {
  autofit?: boolean;
  /**
   * true: wrap at the frame width; false: keep one line. Undefined marks a box
   * from before frames existed, which grows with its text and never clips.
   */
  wrap?: boolean;
  maxFontSize?: number;
  fixedWidth?: number | null;
  fixedHeight?: number | null;
}

export type FrameTextbox = Textbox & { data?: TextBoxData };

/**
 * A text frame keeps the size the user gave it: text wraps (or not) inside it
 * and anything past its edges is cut off, in the editor and in print.
 */
export function isTextFrame(obj: { data?: TextBoxData } | null | undefined): boolean {
  return obj?.data?.wrap !== undefined || !!obj?.data?.autofit;
}

interface TextboxInternals {
  _wrapText(lines: string[], desiredWidth: number): string[][];
  calcTextHeight(): number;
  _render(ctx: CanvasRenderingContext2D): void;
  graphemeSplit(text: string): string[];
}

let installed = false;

/**
 * Fabric's Textbox always wraps and always grows to fit its text. Three small
 * hooks make a frame behave like the renderer: no wrapping when wrap is off,
 * the frame height instead of the text height, and a clip at the frame edges.
 * Boxes without frame data keep Fabric's behaviour.
 */
export function installTextFrameSupport() {
  if (installed) return;
  installed = true;

  const proto = Textbox.prototype as unknown as TextboxInternals;
  const baseWrapText = proto._wrapText;
  const baseCalcTextHeight = proto.calcTextHeight;
  const baseRender = proto._render;

  proto._wrapText = function (this: FrameTextbox & TextboxInternals, lines: string[], desiredWidth: number) {
    if (this.data?.wrap === false) return lines.map(line => this.graphemeSplit(line));
    return baseWrapText.call(this, lines, desiredWidth);
  };

  proto.calcTextHeight = function (this: FrameTextbox & TextboxInternals) {
    const textHeight = baseCalcTextHeight.call(this);
    const frameHeight = this.data?.fixedHeight;
    return isTextFrame(this) && frameHeight && frameHeight > 0 ? frameHeight : textHeight;
  };

  proto._render = function (this: FrameTextbox & TextboxInternals, ctx: CanvasRenderingContext2D) {
    if (!isTextFrame(this)) {
      baseRender.call(this, ctx);
      return;
    }
    ctx.save();
    ctx.beginPath();
    ctx.rect(-this.width / 2, -this.height / 2, this.width, this.height);
    ctx.clip();
    baseRender.call(this, ctx);
    ctx.restore();
  };
}

/**
 * Re-measure a text box after its data changed. Frames show the top and bottom
 * handles so their height can be dragged; a legacy box grows with its text, so
 * those handles would do nothing there.
 */
export function applyTextFrame(obj: FrameTextbox) {
  const frame = isTextFrame(obj);
  obj.setControlsVisibility({ mt: frame, mb: frame });
  obj.initDimensions();
  obj.setCoords();
  // Only data changed, which Fabric does not watch: drop the cached drawing.
  obj.dirty = true;
}
