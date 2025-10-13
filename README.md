# bitly-ingest
This is the interview project for bitly.  

## Design
This project was designed with scale in mind. The provided test file can easily be processed in a single thread, but if the input were scaled up by orders of magnitude, a single threaded solution 
would not be feasible.  The approach I took was to divide the file into configurable chunks and pass start/end file offsets to worker threads, which parse their alloted chunk of the file.
 Once each worker is finished with its assigned chunk, it receives another set of start/end offsets.  This continues until the entire file is processed.

 There are two settings that can be adjusted to tweak the performance: threads and chunk size.  Threads determine how many parallel threads will be used to process the data.  If the 
 number of threads exceeds the number of cores available, the perfomance will drop due to the overhead of managing multiple threads.  Chunk size determines how large the difference 
 between the start and end offsets are.  Larger chunk size will result in faster performance, but more memory consumption.

 ## Running

 The easiest way to run the project is from the repo root:

 `make run`

 This will build and run the project with the default files which were provided with the instructions

To build the project and run with non-default options:

`make build`

This will build the project in `bin/bitly-ingest`

## Testing

To run the included unit tests, run:

`make test`

## Environment Variables

The provided default values will run the project in accordance with the spec, but for test purposes, the following varialbes and their defaults are available:


 ```
THREADS="4"                                 # Number of parallel threads to use to process the decodes file
CHUNK_SIZE_IN_KB="200"                      # Number of bytes each thread reads
START_TIME="2021-01-01T00:00:00Z"           # Only agrigate clicks occuring after this date
END_TIME="2021-12-31T23:59:59Z"             # Only agrigate clicks occuring before this time
ENCODING_FILE_PATH="data/encodes.csv"       # CSV file to read encoding data from
DECODES_FILE_PATH="data/decodes.json"       # File containing the decodes entries to be processed
DEBUG="false"                               # Set the debug flag.  This will dump debug logs
```
