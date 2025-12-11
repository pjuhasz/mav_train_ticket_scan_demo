#define STB_IMAGE_IMPLEMENTATION
#include "stb_image.h"

#include <ZXing/GTIN.h>
#include <ZXing/ReadBarcode.h>
//#include <ZXing/DecodeHints.h>
//#include <ZXing/ImageView.h>
//#include <ZXing/ImageFormat.h>
#include <ZXing/BarcodeFormat.h>

#include <iostream>
#include <iterator>

int main(int argc, char* argv[])
{
	if (argc < 2) {
		std::cerr << "ERROR,Usage: [-b]" << argv[0] << " <image-file>" << std::endl;
		return 1;
	}

	bool bytes_only = false;

	const char* filename;
	
	if (argc >= 3 && strlen(argv[1]) > 1 && strncmp(argv[1], "-b", strlen(argv[1])) == 0) {
		bytes_only = true;
		filename = argv[2];
	} else {
		filename = argv[1];
	}

	// Load image using stb_image as 4‑channel RGBA (8 bits per channel)
	int width = 0, height = 0, channels_in_file = 0;
	unsigned char* pixels = stbi_load(filename, &width, &height, &channels_in_file, 4);
	if (!pixels) {
		std::cerr << "ERROR,Failed to load image: " << filename << std::endl;
		return 1;
	}

	// Configure ZXing decode hints
	ZXing::ReaderOptions options;
	options.setTryHarder(true);
	options.setReturnErrors(true);
	options.setTryRotate(true);

	// Restrict decoding to Aztec (recommended when you know the format).
	// NOTE: Depending on your zxing-cpp version, the enum may be named
	//	   BarcodeFormat::Aztec or BarcodeFormat::AZTEC, and the method
	//	   may be setFormats(...) or setPossibleFormats(...).
	//hints.setFormats(ZXing::BarcodeFormat::Aztec);

	// Wrap the raw pixel buffer with ZXing::ImageView.
	// We told stb_image to give us 4 channels, so this is RGBX.
	ZXing::ImageView imageView(
		pixels,
		width,
		height,
		ZXing::ImageFormat::RGBX
	);

	// Decode the barcodes
	auto barcodes = ZXing::ReadBarcodes(imageView, options);

	// We no longer need the pixel data
	stbi_image_free(pixels);

	for (auto&& barcode : barcodes) {

		if (!barcode.isValid()) {
			std::cerr << "NOT_FOUND";
			if (barcode.error()) {
				std::cerr << "," << ZXing::ToString(barcode.error()) << std::endl;
			}
			continue;
		}

		if (bytes_only) {
			std::cout.write(reinterpret_cast<const char*>(barcode.bytes().data()), barcode.bytes().size());
		} else {
			std::cout << "OK," << ZXing::ToString(barcode.format()) << "," << ZXing::ToHex(barcode.bytes()) << std::endl;
		}

	}

	return 0;
}
