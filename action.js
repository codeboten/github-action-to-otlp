function parseHeaders(headers) {
    if (!headers) {
        return {};
    }
    const parsedHeaders = {};
    for (const header of headers.split(',')) {
        const [key, value] = header.split('=');
        if (!value) {
            continue
        }
        parsedHeaders[key] = value;
    }
    return parsedHeaders;
}

exports.parseHeaders = parseHeaders;