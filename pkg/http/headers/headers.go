package headers

const (
	// ContentType indicates the media type of the data being sent.
	ContentType = "Content-Type"

	// ContentLength indicates the size of the message body, in bytes, sent to the recipient.
	ContentLength = "Content-Length"

	// ContentTypeApplicationJson indicates that the body of the HTTP request or response contains JSON.
	ContentTypeApplicationJson = "application/json"

	// TransferEncoding specifies the form of encoding used to transfer the payload body to the caller.
	TransferEncoding = "Transfer-Encoding"

	// TransferEncodingChunked allows data to be sent in a series of chunks without specifying the total size beforehand.
	TransferEncodingChunked = "chunked"
)
