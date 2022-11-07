const action = require('./action')

test('parseHeader parses correctly', () => {
    expect(action.parseHeaders()).toEqual({});
    expect(action.parseHeaders("key1=value1,key2=value2")).toEqual({
        key1: "value1",
        key2: "value2"
    });
    expect(action.parseHeaders("key1value1,key2=value2")).toEqual({
        key2: "value2"
    });
    expect(action.parseHeaders("key1value1key2=value2")).toEqual({
        key1value1key2: "value2"
    });
});