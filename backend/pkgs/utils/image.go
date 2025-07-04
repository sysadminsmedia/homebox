package utils

import "image"

// flipHorizontal will flip the image horizontally. There is a limit of 10000 pixels in either dimension to prevent excessive memory usage.
func flipHorizontal(img image.Image) image.Image {
	b := img.Bounds()
	if b.Dx() > 10000 || b.Dy() > 10000 {
		return img
	}
	dst := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(b.Max.X-1-(x-b.Min.X), y, img.At(x, y))
		}
	}
	return dst
}

// flipVertical will flip the image vertically. There is a limit of 10000 pixels in either dimension to prevent excessive memory usage.
func flipVertical(img image.Image) image.Image {
	b := img.Bounds()
	if b.Dx() > 10000 || b.Dy() > 10000 {
		return img
	}
	dst := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(x, b.Max.Y-1-(y-b.Min.Y), img.At(x, y))
		}
	}
	return dst
}

// rotate90 will rotate the image 90 degrees clockwise. There is a limit of 10000 pixels in either dimension to prevent excessive memory usage.
func rotate90(img image.Image) image.Image {
	b := img.Bounds()
	if b.Dx() > 10000 || b.Dy() > 10000 {
		return img
	}
	dst := image.NewRGBA(image.Rect(0, 0, b.Dy(), b.Dx()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(b.Max.Y-1-y, x, img.At(x, y))
		}
	}
	return dst
}

func rotate180(img image.Image) image.Image {
	return rotate90(rotate90(img))
}

func rotate270(img image.Image) image.Image {
	return rotate90(rotate180(img))
}

// Applies EXIF orientation using only stdlib
func ApplyOrientation(img image.Image, orientation uint16) image.Image {
	if img == nil {
		return nil
	}
	if orientation < 1 || orientation > 8 {
		return img // No orientation or invalid orientation
	}
	switch orientation {
	case 1:
		return img // No rotation needed
	case 2:
		return flipHorizontal(img)
	case 3:
		return rotate180(img)
	case 4:
		return flipVertical(img)
	case 5:
		return rotate90(flipHorizontal(img))
	case 6:
		return rotate90(img)
	case 7:
		return rotate270(flipHorizontal(img))
	case 8:
		return rotate270(img)
	default:
		return img
	}
}
