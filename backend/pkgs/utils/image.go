package utils

import "image"

// Pure stdlib rotation/flipping
func flipHorizontal(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(b.Max.X-1-(x-b.Min.X), y, img.At(x, y))
		}
	}
	return dst
}

func flipVertical(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(x, b.Max.Y-1-(y-b.Min.Y), img.At(x, y))
		}
	}
	return dst
}

func rotate90(img image.Image) image.Image {
	b := img.Bounds()
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
	switch orientation {
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
