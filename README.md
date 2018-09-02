# Fourier-ADC
Dynamic properties of Analog-Digital Converter calculator written in golang!

```commandline
$ ./Fourier-ADC --help

Starting Fourier-ADC
Usage of Fourier-ADC:
  -dftlen int
        Length of the DFT. (default 1024)
  -fsam float
        Sampling frequency.
  -fsig float
        Original signal frequency.
  -input string
        Input file path.
  -loglevel string
        Logging level. (default "info")

```

Example Output:
```commandline
$ ./Fourier-ADC --dftlen 1024 --input data/example_data.csv --fsig 402.34375 --fsam 4000 --loglevel debug

Starting Fourier-ADC
INFO[0000] Setting logging level debug                  
DEBU[0000] Reading data from data/example_data.csv      
DEBU[0000] Read 1024 points                             
Input signal parameters:
+-----------+----------+---------+---------+
| FSIG [HZ] | FS [HZ]  | DFT LEN | FB [HZ] |
+-----------+----------+---------+---------+
|   402.344 | 4000.000 |    1024 |   3.906 |
+-----------+----------+---------+---------+
DEBU[0000] Calculated DFT in 16.34071ms.                
INFO[0000] DFT plot saved in data/example_data.png      
DEBU[0000] Signal index: 103                            
DEBU[0000] DFT len : 1024                               
DEBU[0000] Aliased harmonics indices [103 206 309 412 509 406 303 200 97 6] 
DEBU[0000] Aliased harmonics freqs [402 805 1.21e+03 1.61e+03 1.99e+03 1.59e+03 1.18e+03 781 379 23.4] 
ADC metrics:
+----------+-----------+-----------+------------+-------------+
| THD [DB] | SNHR [DB] | SFDR [DB] | SINAD [DB] | ENOB [BITS] |
+----------+-----------+-----------+------------+-------------+
|  -72.792 |    50.011 |    65.768 |     49.988 |       8.011 |
+----------+-----------+-----------+------------+-------------+


```
![example_plot](./data/example_data.png)

