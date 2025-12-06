# MÁV train ticket barcode scanner demo

This repository contains a quick and dirty example application that uses
the MÁV train ticket barcode decoder module at [https://github.com/pjuhasz/mav_train_ticket](https://github.com/pjuhasz/mav_train_ticket)
and [ZXing-CPP](https://github.com/zxing-cpp/zxing-cpp) to actually scan
train tickets and show their contents.

## Requirements

- go
- gcc
- make
- the zxing-cpp library and its development headers

## Compilation

Just run `make`.
The binaries will be put in the `bin` directory.

## Running

Run `bin/server`, open `localhost:54623` in your browser, press "Start scanning"
and hold up your ticket to your camera!

The JS code in the frontend sends an image every seconds to the backend,
which runs the separate zxing scanner binary to find the barcode and extract its contents.
If readout was successful the server sends back the contents and the frontend
stops taking images and displays the results.

## Disclaimer

Most of the frontend (if you can call it that) was coded by ChatGPT.
