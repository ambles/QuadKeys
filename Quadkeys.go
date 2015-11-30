// Quadkeys project Quadkeys.go
package Quadkeys

import (
	"bytes"
	"math"
	"strconv"
)

// Source: https://msdn.microsoft.com/en-us/library/bb259689.aspx

const EarthRadius = 6378137
const MinLatitude = -85.05112878
const MaxLatitude = -1 * MinLatitude
const MinLongitude = -180
const MaxLongitude = -1 * MinLongitude
const MaxLevel = 23

func init() {
	// initialization code here
}

/// <summary>
/// Clips a number to the specified minimum and maximum values.
/// </summary>
/// <param name="n">The number to clip.</param>
/// <param name="minValue">Minimum allowable value.</param>
/// <param name="maxValue">Maximum allowable value.</param>
/// <returns>The clipped value.</returns>
func clip(n float64, minValue float64, maxValue float64) float64 {
	return math.Min(math.Max(n, minValue), maxValue)
}

/// <summary>
/// Determines the map width and height (in pixels) at a specified level
/// of detail.
/// </summary>
/// <param name="levelOfDetail">Level of detail, from 1 (lowest detail)
/// to 23 (highest detail).</param>
/// <returns>The map width and height in pixels.</returns>
func MapSize(levelOfDetail uint) uint {
	return 256 << levelOfDetail
}

/// <summary>
/// Determines the ground resolution (in meters per pixel) at a specified
/// latitude and level of detail.
/// </summary>
/// <param name="latitude">Latitude (in degrees) at which to measure the
/// ground resolution.</param>
/// <param name="levelOfDetail">Level of detail, from 1 (lowest detail)
/// to 23 (highest detail).</param>
/// <returns>The ground resolution, in meters per pixel.</returns>
func GroundResolution(latitude float64, levelOfDetail uint) float64 {
	latitude = clip(latitude, MinLatitude, MaxLatitude)
	return math.Cos(latitude*math.Pi/180) * 2 * math.Pi * EarthRadius / float64(MapSize(levelOfDetail))
}

/// <summary>
/// Determines the map scale at a specified latitude, level of detail,
/// and screen resolution.
/// </summary>
/// <param name="latitude">Latitude (in degrees) at which to measure the
/// map scale.</param>
/// <param name="levelOfDetail">Level of detail, from 1 (lowest detail)
/// to 23 (highest detail).</param>
/// <param name="screenDpi">Resolution of the screen, in dots per inch.</param>
/// <returns>The map scale, expressed as the denominator N of the ratio 1 : N.</returns>
func MapScale(latitude float64, levelOfDetail uint, screenDpi uint) float64 {
	return GroundResolution(latitude, levelOfDetail) * float64(screenDpi) / 0.0254
}

/// <summary>
/// Converts a point from latitude/longitude WGS-84 coordinates (in degrees)
/// into pixel XY coordinates at a specified level of detail.
/// </summary>
/// <param name="latitude">Latitude of the point, in degrees.</param>
/// <param name="longitude">Longitude of the point, in degrees.</param>
/// <param name="levelOfDetail">Level of detail, from 1 (lowest detail)
/// to 23 (highest detail).</param>
/// <param name="pixelX">Output parameter receiving the X coordinate in pixels.</param>
/// <param name="pixelY">Output parameter receiving the Y coordinate in pixels.</param>
func LatLongToPixelXY(latitude float64, longitude float64, levelOfDetail uint) (pixelX int, pixelY int) {
	latitude = clip(latitude, MinLatitude, MaxLatitude)
	longitude = clip(longitude, MinLongitude, MaxLongitude)

	x := (longitude + 180) / 360
	sinLatitude := math.Sin(latitude * math.Pi / 180)
	y := 0.5 - math.Log((1+sinLatitude)/(1-sinLatitude))/(4*math.Pi)

	mapSize := MapSize(levelOfDetail)
	pixelX = int(clip(x*float64(mapSize)+0.5, 0, float64(mapSize-1)))
	pixelY = int(clip(y*float64(mapSize)+0.5, 0, float64(mapSize-1)))

	return
}

/// <summary>
/// Converts a pixel from pixel XY coordinates at a specified level of detail
/// into latitude/longitude WGS-84 coordinates (in degrees).
/// </summary>
/// <param name="pixelX">X coordinate of the point, in pixels.</param>
/// <param name="pixelY">Y coordinates of the point, in pixels.</param>
/// <param name="levelOfDetail">Level of detail, from 1 (lowest detail)
/// to 23 (highest detail).</param>
/// <param name="latitude">Output parameter receiving the latitude in degrees.</param>
/// <param name="longitude">Output parameter receiving the longitude in degrees.</param>
func PixelXYToLatLong(pixelX int, pixelY int, levelOfDetail uint) (latitude float64, longitude float64) {
	mapSize := MapSize(levelOfDetail)
	x := (clip(float64(pixelX), 0, float64(mapSize-1)) / float64(mapSize)) - 0.5
	y := 0.5 - (clip(float64(pixelY), 0, float64(mapSize-1)) / float64(mapSize))

	latitude = 90 - 360*math.Atan(math.Exp(-y*2*math.Pi))/math.Pi
	longitude = 360 * x

	return
}

/// <summary>
/// Converts pixel XY coordinates into tile XY coordinates of the tile containing
/// the specified pixel.
/// </summary>
/// <param name="pixelX">Pixel X coordinate.</param>
/// <param name="pixelY">Pixel Y coordinate.</param>
/// <param name="tileX">Output parameter receiving the tile X coordinate.</param>
/// <param name="tileY">Output parameter receiving the tile Y coordinate.</param>
func PixelXYToTileXY(pixelX int, pixelY int) (tileX int, tileY int) {
	tileX = pixelX / 256
	tileY = pixelY / 256
	return
}

/// <summary>
/// Converts tile XY coordinates into pixel XY coordinates of the upper-left pixel
/// of the specified tile.
/// </summary>
/// <param name="tileX">Tile X coordinate.</param>
/// <param name="tileY">Tile Y coordinate.</param>
/// <param name="pixelX">Output parameter receiving the pixel X coordinate.</param>
/// <param name="pixelY">Output parameter receiving the pixel Y coordinate.</param>
func TileXYToPixelXY(tileX int, tileY int) (pixelX int, pixelY int) {
	pixelX = tileX * 256
	pixelY = tileY * 256
	return
}

/// <summary>
/// Converts tile XY coordinates into a QuadKey at a specified level of detail.
/// </summary>
/// <param name="tileX">Tile X coordinate.</param>
/// <param name="tileY">Tile Y coordinate.</param>
/// <param name="levelOfDetail">Level of detail, from 1 (lowest detail)
/// to 23 (highest detail).</param>
/// <returns>A string containing the QuadKey.</returns>
func TileXYToQuadKey(tileX int, tileY int, levelOfDetail uint) string {
	var buffer bytes.Buffer
	for i := levelOfDetail; i > 0; i-- {
		digit := 0
		mask := 1 << (i - 1)
		if (tileX & mask) != 0 {
			digit++
		}
		if (tileY & mask) != 0 {
			digit++
			digit++
		}
		buffer.WriteString(strconv.Itoa(digit))
	}
	return buffer.String()
}

/// <summary>
/// Converts a QuadKey into tile XY coordinates.
/// </summary>
/// <param name="quadKey">QuadKey of the tile.</param>
/// <param name="tileX">Output parameter receiving the tile X coordinate.</param>
/// <param name="tileY">Output parameter receiving the tile Y coordinate.</param>
/// <param name="levelOfDetail">Output parameter receiving the level of detail.</param>
func QuadKeyToTileXY(quadKey string) (tileX int, tileY int, levelOfDetail uint) {
	tileX = 0
	tileY = 0
	levelOfDetail = uint(len(quadKey))
	for i := levelOfDetail; i > 0; i-- {
		mask := 1 << (i - 1)
		switch string(quadKey[levelOfDetail-i]) {
		case "0":

		case "1":
			tileX |= mask

		case "2":
			tileY |= mask

		case "3":
			tileX |= mask
			tileY |= mask

		default:
			tileX = -1
			tileY = -1
		}
	}
	return
}

func LatLongToQuadKey(latitude float64, longitude float64, levelOfDetail uint) string {
	x, y := LatLongToPixelXY(latitude, longitude, levelOfDetail)
	tileX, tileY := PixelXYToTileXY(x, y)
	return TileXYToQuadKey(tileX, tileY, levelOfDetail)
}
