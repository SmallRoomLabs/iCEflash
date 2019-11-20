# iCEflash
A command line tool that uploads iCE40 bitstreams to the flash using a Teensy 3.2 (and probably other Teensys as well).

Currently only the HX1K is suported, but it should be easy to support additional iCE40 models. And it is also only tested on macOS. The serial library used (https://github.com/jacobsa/go-serial) do have suport for both Win, Linux and macOS so it should hopefully work there without a problem.


### Command line arguments
```
iceflash <port> [commands]
        -f              Firmware version (default command)
        -s              Serial number of Flash chip
        -h              Halt/Reset the iCE40
        -g              Go/Start the iCE40
        -c              Check status of CDONE pin of the iCE40
        -e              Erase flash
        -1              Delay 1 second
        -w <filename>   Write flash
        -r <filename>   Read flash
        -t [<filename>] Verify crc32 checksum
```


### Connections between iCE40 dev board and Teensy
```
Programming header                  Pins on the Teensy
Olimex iCE40HX1K-EVB

+-----------------+                 GPIO0  ->  CRESET
|                 |                 GPIO1  ->  CDONE
|   3.3V     GND  |                 GPIO10 ->  SS
|                 |                 GPIO11 ->  SDO
-   RXD     TXD   |                 GPIO12 ->  SDI
                  |                 GPIO13 ->  SCK
   CDONE   CRESET |                 GND    ->  GND
_                 |                 (I've also soldered a
|   SDI     SDO   |                  wire for the +5V on the
|                 |                  back of the dev board
|   SCK     SS    |                  and connected it to VBUS
|                 |                  on the Teensy)
+-----------------+
```
