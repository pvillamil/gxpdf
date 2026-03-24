// Package extractor provides use cases for extracting content from PDF documents.
package extractor

import (
	"fmt"

	"github.com/coregx/gxpdf/internal/encoding"
	"github.com/coregx/gxpdf/internal/models/types"
	"github.com/coregx/gxpdf/internal/parser"
)

// colorSpaceDeviceRGB is the default PDF color space for RGB images.
const colorSpaceDeviceRGB = "DeviceRGB"

// ImageExtractor extracts images from PDF pages.
//
// This is an application service that coordinates image extraction
// from PDF documents using the domain model and infrastructure services.
//
// Example:
//
//	reader, _ := parser.OpenPDF("document.pdf")
//	defer reader.Close()
//
//	extractor := NewImageExtractor(reader)
//	images, _ := extractor.ExtractFromPage(0)
//	for _, img := range images {
//	    img.SaveToFile(fmt.Sprintf("image_%d.jpg", i))
//	}
type ImageExtractor struct {
	reader       *parser.Reader
	dctDecoder   *encoding.DCTDecoder
	flateDecoder *encoding.FlateDecoder
}

// NewImageExtractor creates a new image extractor.
//
// Parameters:
//   - reader: PDF reader providing access to document structure
//
// Returns a configured ImageExtractor ready to extract images.
func NewImageExtractor(reader *parser.Reader) *ImageExtractor {
	return &ImageExtractor{
		reader:       reader,
		dctDecoder:   encoding.NewDCTDecoder(),
		flateDecoder: encoding.NewFlateDecoder(),
	}
}

// ExtractFromDocument extracts all images from all pages in the document.
//
// This iterates through all pages and extracts images from each.
//
// Returns a slice of all images found in the document, or error if extraction fails.
func (e *ImageExtractor) ExtractFromDocument() ([]*types.Image, error) {
	pageCount, err := e.reader.GetPageCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get page count: %w", err)
	}

	var allImages []*types.Image

	for i := 0; i < pageCount; i++ {
		images, err := e.ExtractFromPage(i)
		if err != nil {
			// Log error but continue with other pages
			continue
		}
		allImages = append(allImages, images...)
	}

	return allImages, nil
}

// ExtractFromPage extracts all images from a specific page.
//
// This finds all image XObjects in the page's resources and extracts them.
//
// Parameters:
//   - pageIndex: 0-based page index
//
// Returns a slice of images found on the page, or error if extraction fails.
func (e *ImageExtractor) ExtractFromPage(pageIndex int) ([]*types.Image, error) {
	// Get page dictionary
	pageDict, err := e.reader.GetPage(pageIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get page %d: %w", pageIndex, err)
	}

	// Get page resources
	resourcesObj := pageDict.Get("Resources")
	if resourcesObj == nil {
		// No resources means no images
		return []*types.Image{}, nil
	}

	resourcesDict, ok := resourcesObj.(*parser.Dictionary)
	if !ok {
		// Try to resolve indirect reference
		if ref, ok := resourcesObj.(*parser.IndirectReference); ok {
			resolvedObj, err := e.reader.GetObject(ref.Number)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve resources reference: %w", err)
			}
			resourcesDict, ok = resolvedObj.(*parser.Dictionary)
			if !ok {
				return nil, fmt.Errorf("resources is not a dictionary: %T", resolvedObj)
			}
		} else {
			return nil, fmt.Errorf("resources is not a dictionary: %T", resourcesObj)
		}
	}

	// Get XObject dictionary from resources
	xobjectObj := resourcesDict.Get("XObject")
	if xobjectObj == nil {
		// No XObjects means no images
		return []*types.Image{}, nil
	}

	xobjectDict, ok := xobjectObj.(*parser.Dictionary)
	if !ok {
		// Try to resolve indirect reference
		if ref, ok := xobjectObj.(*parser.IndirectReference); ok {
			resolvedObj, err := e.reader.GetObject(ref.Number)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve XObject reference: %w", err)
			}
			xobjectDict, ok = resolvedObj.(*parser.Dictionary)
			if !ok {
				return nil, fmt.Errorf("XObject is not a dictionary: %T", resolvedObj)
			}
		} else {
			return nil, fmt.Errorf("XObject is not a dictionary: %T", xobjectObj)
		}
	}

	// Extract images from XObject dictionary
	var images []*types.Image

	// Iterate over all entries in the XObject dictionary
	keys := xobjectDict.Keys()

	for _, name := range keys {
		xobjRef := xobjectDict.Get(name)
		if xobjRef == nil {
			continue
		}

		// Resolve XObject reference
		var xobj parser.PdfObject
		if ref, ok := xobjRef.(*parser.IndirectReference); ok {
			resolvedObj, err := e.reader.GetObject(ref.Number)
			if err != nil {
				continue // Skip this XObject on error
			}
			xobj = resolvedObj
		} else {
			xobj = xobjRef
		}

		// Check if this XObject is an image
		stream, ok := xobj.(*parser.Stream)
		if !ok {
			continue // Not a stream, skip
		}

		subtypeObj := stream.Dictionary().Get("Subtype")
		if subtypeObj == nil {
			continue
		}

		subtype, ok := subtypeObj.(*parser.Name)
		if !ok || subtype.Value() != "Image" {
			continue // Not an image, skip
		}

		// Extract image from stream
		img, err := e.extractImageFromStream(stream, name)
		if err != nil {
			// Log error but continue with other images
			continue
		}

		images = append(images, img)
	}

	return images, nil
}

// extractImageFromStream extracts an image from a PDF stream object.
//
// This handles different image encodings and color spaces.
func (e *ImageExtractor) extractImageFromStream(stream *parser.Stream, name string) (*types.Image, error) {
	// Get image properties
	dict := stream.Dictionary()
	width := int(dict.GetInteger("Width"))
	height := int(dict.GetInteger("Height"))
	bitsPerComponent := int(dict.GetInteger("BitsPerComponent"))

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid image dimensions: %dx%d", width, height)
	}

	if bitsPerComponent <= 0 {
		bitsPerComponent = 8 // Default to 8 bits
	}

	// Get color space
	colorSpaceObj := dict.Get("ColorSpace")
	colorSpace := e.getColorSpaceName(colorSpaceObj)

	// Get filter
	filterObj := dict.Get("Filter")
	filter := e.getFilterName(filterObj)

	// Decode stream data
	data, err := e.decodeImageData(stream, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image data: %w", err)
	}

	// Create Image value object
	img, err := types.NewImage(data, width, height, colorSpace, bitsPerComponent, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to create image: %w", err)
	}

	// Set name
	img.SetName(name)

	return img, nil
}

// decodeImageData decodes image stream data based on the filter.
func (e *ImageExtractor) decodeImageData(stream *parser.Stream, filter string) ([]byte, error) {
	switch filter {
	case "/DCTDecode":
		// For JPEG, return the raw stream data (already compressed)
		return stream.Content(), nil

	case "/FlateDecode":
		// Decompress using Flate decoder
		rawData := stream.Content()
		decodedData, err := e.flateDecoder.Decode(rawData)
		if err != nil {
			return nil, fmt.Errorf("flate decode failed: %w", err)
		}
		return decodedData, nil

	case "":
		// No filter, return raw data
		return stream.Content(), nil

	default:
		return nil, fmt.Errorf("unsupported filter: %s", filter)
	}
}

// getColorSpaceName extracts the color space name from a PDF object.
func (e *ImageExtractor) getColorSpaceName(obj parser.PdfObject) string {
	if obj == nil {
		return colorSpaceDeviceRGB // Default color space
	}

	// Direct name (e.g., /DeviceRGB)
	if name, ok := obj.(*parser.Name); ok {
		return name.Value()
	}

	// Array (e.g., [/Indexed ...])
	if arr, ok := obj.(*parser.Array); ok {
		if arr.Len() > 0 {
			if name, ok := arr.Get(0).(*parser.Name); ok {
				return name.Value()
			}
		}
	}

	return colorSpaceDeviceRGB // Default
}

// getFilterName extracts the filter name from a PDF object.
func (e *ImageExtractor) getFilterName(obj parser.PdfObject) string {
	if obj == nil {
		return "" // No filter
	}

	// Direct name (e.g., /DCTDecode)
	if name, ok := obj.(*parser.Name); ok {
		return name.Value()
	}

	// Array of filters (use first filter)
	if arr, ok := obj.(*parser.Array); ok {
		if arr.Len() > 0 {
			if name, ok := arr.Get(0).(*parser.Name); ok {
				return name.Value()
			}
		}
	}

	return "" // No filter
}
