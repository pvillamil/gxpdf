// Package main demonstrates gradient fills in PDF creation.
//
// This example shows how to use linear and radial gradients to fill shapes
// using PDF Shading dictionaries (ShadingType 2 for linear, ShadingType 3 for radial).
package main

import (
	"fmt"
	"log"

	"github.com/coregx/gxpdf/creator"
)

func main() {
	// Create PDF creator
	c := creator.New()
	c.SetTitle("Gradient Fills Example")
	c.SetAuthor("GxPDF")

	// Create a new page
	page, err := c.NewPage()
	if err != nil {
		log.Fatalf("Failed to create page: %v", err)
	}

	// Add title
	err = page.AddText("Gradient Fills Example", 50, 750, creator.HelveticaBold, 24)
	if err != nil {
		log.Fatalf("Failed to add title: %v", err)
	}

	// Example 1: Linear gradient rectangle (horizontal)
	err = page.AddText("1. Linear Gradient (Horizontal)", 50, 700, creator.Helvetica, 12)
	if err != nil {
		log.Fatalf("Failed to add text: %v", err)
	}

	linearH := creator.NewLinearGradient(50, 650, 250, 650) // Left to right
	linearH.AddColorStop(0, creator.Red)
	linearH.AddColorStop(0.5, creator.Yellow)
	linearH.AddColorStop(1, creator.Green)

	err = page.DrawRect(50, 620, 200, 60, &creator.RectOptions{
		FillGradient: linearH,
		StrokeColor:  &creator.Black,
		StrokeWidth:  1,
	})
	if err != nil {
		log.Fatalf("Failed to draw linear gradient rectangle: %v", err)
	}

	// Example 2: Linear gradient rectangle (vertical)
	err = page.AddText("2. Linear Gradient (Vertical)", 300, 700, creator.Helvetica, 12)
	if err != nil {
		log.Fatalf("Failed to add text: %v", err)
	}

	linearV := creator.NewLinearGradient(350, 620, 350, 680) // Bottom to top
	linearV.AddColorStop(0, creator.Blue)
	linearV.AddColorStop(1, creator.White)

	err = page.DrawRect(300, 620, 200, 60, &creator.RectOptions{
		FillGradient: linearV,
		StrokeColor:  &creator.Black,
		StrokeWidth:  1,
	})
	if err != nil {
		log.Fatalf("Failed to draw vertical gradient rectangle: %v", err)
	}

	// Example 3: Radial gradient circle
	err = page.AddText("3. Radial Gradient (Circle)", 50, 570, creator.Helvetica, 12)
	if err != nil {
		log.Fatalf("Failed to add text: %v", err)
	}

	radial := creator.NewRadialGradient(150, 480, 0, 150, 480, 60) // Center to edge
	radial.AddColorStop(0, creator.White)
	radial.AddColorStop(0.5, creator.Yellow)
	radial.AddColorStop(1, creator.Red)

	err = page.DrawCircle(150, 480, 60, &creator.CircleOptions{
		FillGradient: radial,
		StrokeColor:  &creator.Black,
		StrokeWidth:  2,
	})
	if err != nil {
		log.Fatalf("Failed to draw radial gradient circle: %v", err)
	}

	// Example 4: Linear gradient polygon (star shape)
	err = page.AddText("4. Linear Gradient (Polygon)", 300, 570, creator.Helvetica, 12)
	if err != nil {
		log.Fatalf("Failed to add text: %v", err)
	}

	linearDiag := creator.NewLinearGradient(300, 420, 450, 540) // Diagonal
	linearDiag.AddColorStop(0, creator.Color{R: 1, G: 0, B: 1}) // Magenta
	linearDiag.AddColorStop(1, creator.Color{R: 0, G: 1, B: 1}) // Cyan

	// Create a star polygon
	starVertices := []creator.Point{
		{X: 400, Y: 540}, // Top
		{X: 420, Y: 490},
		{X: 470, Y: 490},
		{X: 430, Y: 460},
		{X: 450, Y: 410}, // Right
		{X: 400, Y: 440},
		{X: 350, Y: 410}, // Left
		{X: 370, Y: 460},
		{X: 330, Y: 490},
		{X: 380, Y: 490},
	}

	err = page.DrawPolygon(starVertices, &creator.PolygonOptions{
		FillGradient: linearDiag,
		StrokeColor:  &creator.Black,
		StrokeWidth:  1.5,
	})
	if err != nil {
		log.Fatalf("Failed to draw gradient polygon: %v", err)
	}

	// Example 5: Ellipse with radial gradient
	err = page.AddText("5. Radial Gradient (Ellipse)", 50, 370, creator.Helvetica, 12)
	if err != nil {
		log.Fatalf("Failed to add text: %v", err)
	}

	radialEllipse := creator.NewRadialGradient(150, 280, 0, 150, 280, 80)
	radialEllipse.AddColorStop(0, creator.Color{R: 1, G: 1, B: 0.8}) // Light yellow
	radialEllipse.AddColorStop(1, creator.Color{R: 1, G: 0.5, B: 0}) // Orange

	err = page.DrawEllipse(150, 280, 100, 60, &creator.EllipseOptions{
		FillGradient: radialEllipse,
		StrokeColor:  &creator.Black,
		StrokeWidth:  1,
	})
	if err != nil {
		log.Fatalf("Failed to draw gradient ellipse: %v", err)
	}

	// Add footer note
	err = page.AddText("Gradients rendered using PDF Shading dictionaries (Type 2 + Type 3).",
		50, 50, creator.Helvetica, 10)
	if err != nil {
		log.Fatalf("Failed to add footer: %v", err)
	}

	// Write PDF to file
	err = c.WriteToFile("gradients_example.pdf")
	if err != nil {
		log.Fatalf("Failed to write PDF: %v", err)
	}

	fmt.Println("PDF created successfully: gradients_example.pdf")
	fmt.Println("\nGradient fills have been added to:")
	fmt.Println("  - Rectangles (horizontal and vertical)")
	fmt.Println("  - Circles (radial gradient)")
	fmt.Println("  - Polygons (diagonal linear gradient)")
	fmt.Println("  - Ellipses (radial gradient)")
}
