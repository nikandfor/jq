/*
jq is unstructured data manipulation library. It cat decode input in one format, process it, and encode it into another format.

At first get familiar with original jq language: https://jqlang.github.io/jq/.
The library is aimed to be almost fully compatible with original jq.

The central concept here is Filter. It works on data stored in Buffer.
At first data is usually decoded from one of the supported encodings: json, csv, url-encoding, cbor, base64, or implement your own.
Then combination of filters is used to transform data in the Buffer.
And then data is used directly or encoded into one of the supported formats.
Sandwich is a helper type combining this tasks into a single call.

Filters are stateful, so they can't be used concurrently for different pipelines.
But they can be reused in a sequential manner, ie process one value, then the next one, and so on.

Filter takes one value and produces zero, one, or more values returning at most one at a time.
The simplest example is an array iterator.
It takes an array of arbitrary size and returns its elements one at a time.

Data is stored as a tree of nodes in an append only data structure (Buffer),
filters take and return references, technically offsets in the buffer.
Basic values are stored once and then references to it are passed between filters if needed.
Arrays and Maps are compact lists of references to it's elements.
Some frequently used values like Null, True, False, 0, 1 are embedded into the Off, so they don't take space in the Buffer.

Buffer.Reset should be called when processong is done to reclaim memory.
All the Off references become invalid at that moment except of embedded ones.

Data Decoders/Encoders for different encodings are in directories ./jqXXX.
*/
package jq
