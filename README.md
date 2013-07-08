#go-bmp085

*A port of [Adafruit's Python library](https://github.com/adafruit/Adafruit-Raspberry-Pi-Python-Code/tree/master/Adafruit_BMP085) for [BMP05](http://adafru.it/391) to Go.*


## Installation

<pre>
$ go get github.com/stratoberry/go-bmp085
</pre>

[i2c library by gmcbay](http://bitbucket.org/gmcbay/i2c) is the only dependency.

## Usage

Library is similar to the one by Adafruit but with a few minor changes to match Go's conventions.

<pre><code>dev, _ := bmp085.Init(0x77, 0, bmp085.MODE_STANDARD)
temperate, pressure, altitude, _ = dev.GetData()
</code></pre>

For a full example take a look at the [examples/main.go](https://github.com/stratoberry/go-bmp085/blob/master/examples/main.go) file in the repository.
