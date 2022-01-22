/*! osc.js 2.4.1, Copyright 2021 Colin Clark | github.com/colinbdclark/osc.js */

/*
 * osc.js: An Open Sound Control library for JavaScript that works in both the browser and Node.js
 *
 * Copyright 2014-2016, Colin Clark
 * Licensed under the MIT and GPL 3 licenses.
 */

/* global require, module, process, Buffer, Long, util */

var osc = osc || {};

(function () {

    "use strict";

    osc.SECS_70YRS = 2208988800;
    osc.TWO_32 = 4294967296;

    osc.defaults = {
        metadata: false,
        unpackSingleArgs: true
    };

    // Unsupported, non-API property.
    osc.isCommonJS = typeof module !== "undefined" && module.exports ? true : false;

    // Unsupported, non-API property.
    osc.isNode = osc.isCommonJS && typeof window === "undefined";

    // Unsupported, non-API property.
    osc.isElectron = typeof process !== "undefined" &&
        process.versions && process.versions.electron ? true : false;

    // Unsupported, non-API property.
    osc.isBufferEnv = osc.isNode || osc.isElectron;

    // Unsupported, non-API function.
    osc.isArray = function (obj) {
        return obj && Object.prototype.toString.call(obj) === "[object Array]";
    };

    // Unsupported, non-API function.
    osc.isTypedArrayView = function (obj) {
        return obj.buffer && obj.buffer instanceof ArrayBuffer;
    };

    // Unsupported, non-API function.
    osc.isBuffer = function (obj) {
        return osc.isBufferEnv && obj instanceof Buffer;
    };

    // Unsupported, non-API member.
    osc.Long = typeof Long !== "undefined" ? Long :
        osc.isNode ? require("long") : undefined;

    // Unsupported, non-API member. Can be removed when supported versions
    // of Node.js expose TextDecoder as a global, as in the browser.
    osc.TextDecoder = typeof TextDecoder !== "undefined" ? new TextDecoder("utf-8") :
        typeof util !== "undefined" && typeof (util.TextDecoder !== "undefined") ? new util.TextDecoder("utf-8") : undefined;

    osc.TextEncoder = typeof TextEncoder !== "undefined" ? new TextEncoder("utf-8") :
        typeof util !== "undefined" && typeof (util.TextEncoder !== "undefined") ? new util.TextEncoder("utf-8") : undefined;

    /**
     * Wraps the specified object in a DataView.
     *
     * @param {Array-like} obj the object to wrap in a DataView instance
     * @return {DataView} the DataView object
     */
    // Unsupported, non-API function.
    osc.dataView = function (obj, offset, length) {
        if (obj.buffer) {
            return new DataView(obj.buffer, offset, length);
        }

        if (obj instanceof ArrayBuffer) {
            return new DataView(obj, offset, length);
        }

        return new DataView(new Uint8Array(obj), offset, length);
    };

    /**
     * Takes an ArrayBuffer, TypedArray, DataView, Buffer, or array-like object
     * and returns a Uint8Array view of it.
     *
     * Throws an error if the object isn't suitably array-like.
     *
     * @param {Array-like or Array-wrapping} obj an array-like or array-wrapping object
     * @returns {Uint8Array} a typed array of octets
     */
    // Unsupported, non-API function.
    osc.byteArray = function (obj) {
        if (obj instanceof Uint8Array) {
            return obj;
        }

        var buf = obj.buffer ? obj.buffer : obj;

        if (!(buf instanceof ArrayBuffer) && (typeof buf.length === "undefined" || typeof buf === "string")) {
            throw new Error("Can't wrap a non-array-like object as Uint8Array. Object was: " +
                JSON.stringify(obj, null, 2));
        }


        // TODO gh-39: This is a potentially unsafe algorithm;
        // if we're getting anything other than a TypedArrayView (such as a DataView),
        // we really need to determine the range of the view it is viewing.
        return new Uint8Array(buf);
    };

    /**
     * Takes an ArrayBuffer, TypedArray, DataView, or array-like object
     * and returns a native buffer object
     * (i.e. in Node.js, a Buffer object and in the browser, a Uint8Array).
     *
     * Throws an error if the object isn't suitably array-like.
     *
     * @param {Array-like or Array-wrapping} obj an array-like or array-wrapping object
     * @returns {Buffer|Uint8Array} a buffer object
     */
    // Unsupported, non-API function.
    osc.nativeBuffer = function (obj) {
        if (osc.isBufferEnv) {
            return osc.isBuffer(obj) ? obj :
                Buffer.from(obj.buffer ? obj : new Uint8Array(obj));
        }

        return osc.isTypedArrayView(obj) ? obj : new Uint8Array(obj);
    };

    // Unsupported, non-API function
    osc.copyByteArray = function (source, target, offset) {
        if (osc.isTypedArrayView(source) && osc.isTypedArrayView(target)) {
            target.set(source, offset);
        } else {
            var start = offset === undefined ? 0 : offset,
                len = Math.min(target.length - offset, source.length);

            for (var i = 0, j = start; i < len; i++, j++) {
                target[j] = source[i];
            }
        }

        return target;
    };

    /**
     * Reads an OSC-formatted string.
     *
     * @param {DataView} dv a DataView containing the raw bytes of the OSC string
     * @param {Object} offsetState an offsetState object used to store the current offset index
     * @return {String} the JavaScript String that was read
     */
    osc.readString = function (dv, offsetState) {
        var charCodes = [],
            idx = offsetState.idx;

        for (; idx < dv.byteLength; idx++) {
            var charCode = dv.getUint8(idx);
            if (charCode !== 0) {
                charCodes.push(charCode);
            } else {
                idx++;
                break;
            }
        }

        // Round to the nearest 4-byte block.
        idx = (idx + 3) & ~0x03;
        offsetState.idx = idx;

        var decoder = osc.isBufferEnv ? osc.readString.withBuffer :
            osc.TextDecoder ? osc.readString.withTextDecoder : osc.readString.raw;

        return decoder(charCodes);
    };

    osc.readString.raw = function (charCodes) {
        // If no Buffer or TextDecoder, resort to fromCharCode
        // This does not properly decode multi-byte Unicode characters.
        var str = "";
        var sliceSize = 10000;

        // Processing the array in chunks so as not to exceed argument
        // limit, see https://bugs.webkit.org/show_bug.cgi?id=80797
        for (var i = 0; i < charCodes.length; i += sliceSize) {
            str += String.fromCharCode.apply(null, charCodes.slice(i, i + sliceSize));
        }

        return str;
    };

    osc.readString.withTextDecoder = function (charCodes) {
        var data = new Int8Array(charCodes);
        return osc.TextDecoder.decode(data);
    };

    osc.readString.withBuffer = function (charCodes) {
        return Buffer.from(charCodes).toString("utf-8");
    };

    /**
     * Writes a JavaScript string as an OSC-formatted string.
     *
     * @param {String} str the string to write
     * @return {Uint8Array} a buffer containing the OSC-formatted string
     */
    osc.writeString = function (str) {
        var terminated = str + "\u0000",
            len = terminated.length,
            paddedLen = (len + 3) & ~0x03,
            arr = new Uint8Array(paddedLen);

        var encoder = osc.isBufferEnv ? osc.writeString.withBuffer :
            osc.TextEncoder ? osc.writeString.withTextEncoder : null,
            encodedStr;

        if (encoder) {
            encodedStr = encoder(terminated);
        }

        for (var i = 0; i < terminated.length; i++) {
            var charCode = encoder ? encodedStr[i] : terminated.charCodeAt(i);
            arr[i] = charCode;
        }

        return arr;
    };

    osc.writeString.withTextEncoder = function (str) {
        return osc.TextEncoder.encode(str);
    };

    osc.writeString.withBuffer = function (str) {
        return Buffer.from(str);
    };

    // Unsupported, non-API function.
    osc.readPrimitive = function (dv, readerName, numBytes, offsetState) {
        var val = dv[readerName](offsetState.idx, false);
        offsetState.idx += numBytes;

        return val;
    };

    // Unsupported, non-API function.
    osc.writePrimitive = function (val, dv, writerName, numBytes, offset) {
        offset = offset === undefined ? 0 : offset;

        var arr;
        if (!dv) {
            arr = new Uint8Array(numBytes);
            dv = new DataView(arr.buffer);
        } else {
            arr = new Uint8Array(dv.buffer);
        }

        dv[writerName](offset, val, false);

        return arr;
    };

    /**
     * Reads an OSC int32 ("i") value.
     *
     * @param {DataView} dv a DataView containing the raw bytes
     * @param {Object} offsetState an offsetState object used to store the current offset index into dv
     * @return {Number} the number that was read
     */
    osc.readInt32 = function (dv, offsetState) {
        return osc.readPrimitive(dv, "getInt32", 4, offsetState);
    };

    /**
     * Writes an OSC int32 ("i") value.
     *
     * @param {Number} val the number to write
     * @param {DataView} [dv] a DataView instance to write the number into
     * @param {Number} [offset] an offset into dv
     */
    osc.writeInt32 = function (val, dv, offset) {
        return osc.writePrimitive(val, dv, "setInt32", 4, offset);
    };

    /**
     * Reads an OSC int64 ("h") value.
     *
     * @param {DataView} dv a DataView containing the raw bytes
     * @param {Object} offsetState an offsetState object used to store the current offset index into dv
     * @return {Number} the number that was read
     */
    osc.readInt64 = function (dv, offsetState) {
        var high = osc.readPrimitive(dv, "getInt32", 4, offsetState),
            low = osc.readPrimitive(dv, "getInt32", 4, offsetState);

        if (osc.Long) {
            return new osc.Long(low, high);
        } else {
            return {
                high: high,
                low: low,
                unsigned: false
            };
        }
    };

    /**
     * Writes an OSC int64 ("h") value.
     *
     * @param {Number} val the number to write
     * @param {DataView} [dv] a DataView instance to write the number into
     * @param {Number} [offset] an offset into dv
     */
    osc.writeInt64 = function (val, dv, offset) {
        var arr = new Uint8Array(8);
        arr.set(osc.writePrimitive(val.high, dv, "setInt32", 4, offset), 0);
        arr.set(osc.writePrimitive(val.low,  dv, "setInt32", 4, offset + 4), 4);
        return arr;
    };

    /**
     * Reads an OSC float32 ("f") value.
     *
     * @param {DataView} dv a DataView containing the raw bytes
     * @param {Object} offsetState an offsetState object used to store the current offset index into dv
     * @return {Number} the number that was read
     */
    osc.readFloat32 = function (dv, offsetState) {
        return osc.readPrimitive(dv, "getFloat32", 4, offsetState);
    };

    /**
     * Writes an OSC float32 ("f") value.
     *
     * @param {Number} val the number to write
     * @param {DataView} [dv] a DataView instance to write the number into
     * @param {Number} [offset] an offset into dv
     */
    osc.writeFloat32 = function (val, dv, offset) {
        return osc.writePrimitive(val, dv, "setFloat32", 4, offset);
    };

    /**
     * Reads an OSC float64 ("d") value.
     *
     * @param {DataView} dv a DataView containing the raw bytes
     * @param {Object} offsetState an offsetState object used to store the current offset index into dv
     * @return {Number} the number that was read
     */
    osc.readFloat64 = function (dv, offsetState) {
        return osc.readPrimitive(dv, "getFloat64", 8, offsetState);
    };

    /**
     * Writes an OSC float64 ("d") value.
     *
     * @param {Number} val the number to write
     * @param {DataView} [dv] a DataView instance to write the number into
     * @param {Number} [offset] an offset into dv
     */
    osc.writeFloat64 = function (val, dv, offset) {
        return osc.writePrimitive(val, dv, "setFloat64", 8, offset);
    };

    /**
     * Reads an OSC 32-bit ASCII character ("c") value.
     *
     * @param {DataView} dv a DataView containing the raw bytes
     * @param {Object} offsetState an offsetState object used to store the current offset index into dv
     * @return {String} a string containing the read character
     */
    osc.readChar32 = function (dv, offsetState) {
        var charCode = osc.readPrimitive(dv, "getUint32", 4, offsetState);
        return String.fromCharCode(charCode);
    };

    /**
     * Writes an OSC 32-bit ASCII character ("c") value.
     *
     * @param {String} str the string from which the first character will be written
     * @param {DataView} [dv] a DataView instance to write the character into
     * @param {Number} [offset] an offset into dv
     * @return {String} a string containing the read character
     */
    osc.writeChar32 = function (str, dv, offset) {
        var charCode = str.charCodeAt(0);
        if (charCode === undefined || charCode < -1) {
            return undefined;
        }

        return osc.writePrimitive(charCode, dv, "setUint32", 4, offset);
    };

    /**
     * Reads an OSC blob ("b") (i.e. a Uint8Array).
     *
     * @param {DataView} dv a DataView instance to read from
     * @param {Object} offsetState an offsetState object used to store the current offset index into dv
     * @return {Uint8Array} the data that was read
     */
    osc.readBlob = function (dv, offsetState) {
        var len = osc.readInt32(dv, offsetState),
            paddedLen = (len + 3) & ~0x03,
            blob = new Uint8Array(dv.buffer, offsetState.idx, len);

        offsetState.idx += paddedLen;

        return blob;
    };

    /**
     * Writes a raw collection of bytes to a new ArrayBuffer.
     *
     * @param {Array-like} data a collection of octets
     * @return {ArrayBuffer} a buffer containing the OSC-formatted blob
     */
    osc.writeBlob = function (data) {
        data = osc.byteArray(data);

        var len = data.byteLength,
            paddedLen = (len + 3) & ~0x03,
            offset = 4, // Extra 4 bytes is for the size.
            blobLen = paddedLen + offset,
            arr = new Uint8Array(blobLen),
            dv = new DataView(arr.buffer);

        // Write the size.
        osc.writeInt32(len, dv);

        // Since we're writing to a real ArrayBuffer,
        // we don't need to pad the remaining bytes.
        arr.set(data, offset);

        return arr;
    };

    /**
     * Reads an OSC 4-byte MIDI message.
     *
     * @param {DataView} dv the DataView instance to read from
     * @param {Object} offsetState an offsetState object used to store the current offset index into dv
     * @return {Uint8Array} an array containing (in order) the port ID, status, data1 and data1 bytes
     */
    osc.readMIDIBytes = function (dv, offsetState) {
        var midi = new Uint8Array(dv.buffer, offsetState.idx, 4);
        offsetState.idx += 4;

        return midi;
    };

    /**
     * Writes an OSC 4-byte MIDI message.
     *
     * @param {Array-like} bytes a 4-element array consisting of the port ID, status, data1 and data1 bytes
     * @return {Uint8Array} the written message
     */
    osc.writeMIDIBytes = function (bytes) {
        bytes = osc.byteArray(bytes);

        var arr = new Uint8Array(4);
        arr.set(bytes);

        return arr;
    };

    /**
     * Reads an OSC RGBA colour value.
     *
     * @param {DataView} dv the DataView instance to read from
     * @param {Object} offsetState an offsetState object used to store the current offset index into dv
     * @return {Object} a colour object containing r, g, b, and a properties
     */
    osc.readColor = function (dv, offsetState) {
        var bytes = new Uint8Array(dv.buffer, offsetState.idx, 4),
            alpha = bytes[3] / 255;

        offsetState.idx += 4;

        return {
            r: bytes[0],
            g: bytes[1],
            b: bytes[2],
            a: alpha
        };
    };

    /**
     * Writes an OSC RGBA colour value.
     *
     * @param {Object} color a colour object containing r, g, b, and a properties
     * @return {Uint8Array} a byte array containing the written color
     */
    osc.writeColor = function (color) {
        var alpha = Math.round(color.a * 255),
            arr = new Uint8Array([color.r, color.g, color.b, alpha]);

        return arr;
    };

    /**
     * Reads an OSC true ("T") value by directly returning the JavaScript Boolean "true".
     */
    osc.readTrue = function () {
        return true;
    };

    /**
     * Reads an OSC false ("F") value by directly returning the JavaScript Boolean "false".
     */
    osc.readFalse = function () {
        return false;
    };

    /**
     * Reads an OSC nil ("N") value by directly returning the JavaScript "null" value.
     */
    osc.readNull = function () {
        return null;
    };

    /**
     * Reads an OSC impulse/bang/infinitum ("I") value by directly returning 1.0.
     */
    osc.readImpulse = function () {
        return 1.0;
    };

    /**
     * Reads an OSC time tag ("t").
     *
     * @param {DataView} dv the DataView instance to read from
     * @param {Object} offsetState an offset state object containing the current index into dv
     * @param {Object} a time tag object containing both the raw NTP as well as the converted native (i.e. JS/UNIX) time
     */
    osc.readTimeTag = function (dv, offsetState) {
        var secs1900 = osc.readPrimitive(dv, "getUint32", 4, offsetState),
            frac = osc.readPrimitive(dv, "getUint32", 4, offsetState),
            native = (secs1900 === 0 && frac === 1) ? Date.now() : osc.ntpToJSTime(secs1900, frac);

        return {
            raw: [secs1900, frac],
            native: native
        };
    };

    /**
     * Writes an OSC time tag ("t").
     *
     * Takes, as its argument, a time tag object containing either a "raw" or "native property."
     * The raw timestamp must conform to the NTP standard representation, consisting of two unsigned int32
     * values. The first represents the number of seconds since January 1, 1900; the second, fractions of a second.
     * "Native" JavaScript timestamps are specified as a Number representing milliseconds since January 1, 1970.
     *
     * @param {Object} timeTag time tag object containing either a native JS timestamp (in ms) or a NTP timestamp pair
     * @return {Uint8Array} raw bytes for the written time tag
     */
    osc.writeTimeTag = function (timeTag) {
        var raw = timeTag.raw ? timeTag.raw : osc.jsToNTPTime(timeTag.native),
            arr = new Uint8Array(8), // Two Unit32s.
            dv = new DataView(arr.buffer);

        osc.writeInt32(raw[0], dv, 0);
        osc.writeInt32(raw[1], dv, 4);

        return arr;
    };

    /**
     * Produces a time tag containing a raw NTP timestamp
     * relative to now by the specified number of seconds.
     *
     * @param {Number} secs the number of seconds relative to now (i.e. + for the future, - for the past)
     * @param {Number} now the number of milliseconds since epoch to use as the current time. Defaults to Date.now()
     * @return {Object} the time tag
     */
    osc.timeTag = function (secs, now) {
        secs = secs || 0;
        now = now || Date.now();

        var nowSecs = now / 1000,
            nowWhole = Math.floor(nowSecs),
            nowFracs = nowSecs - nowWhole,
            secsWhole = Math.floor(secs),
            secsFracs = secs - secsWhole,
            fracs = nowFracs + secsFracs;

        if (fracs > 1) {
            var fracsWhole = Math.floor(fracs),
                fracsFracs = fracs - fracsWhole;

            secsWhole += fracsWhole;
            fracs = fracsFracs;
        }

        var ntpSecs = nowWhole + secsWhole + osc.SECS_70YRS,
            ntpFracs = Math.round(osc.TWO_32 * fracs);

        return {
            raw: [ntpSecs, ntpFracs]
        };
    };

    /**
     * Converts OSC's standard time tag representation (which is the NTP format)
     * into the JavaScript/UNIX format in milliseconds.
     *
     * @param {Number} secs1900 the number of seconds since 1900
     * @param {Number} frac the number of fractions of a second (between 0 and 2^32)
     * @return {Number} a JavaScript-compatible timestamp in milliseconds
     */
    osc.ntpToJSTime = function (secs1900, frac) {
        var secs1970 = secs1900 - osc.SECS_70YRS,
            decimals = frac / osc.TWO_32,
            msTime = (secs1970 + decimals) * 1000;

        return msTime;
    };

    osc.jsToNTPTime = function (jsTime) {
        var secs = jsTime / 1000,
            secsWhole = Math.floor(secs),
            secsFrac = secs - secsWhole,
            ntpSecs = secsWhole + osc.SECS_70YRS,
            ntpFracs = Math.round(osc.TWO_32 * secsFrac);

        return [ntpSecs, ntpFracs];
    };

    /**
     * Reads the argument portion of an OSC message.
     *
     * @param {DataView} dv a DataView instance to read from
     * @param {Object} offsetState the offsetState object that stores the current offset into dv
     * @param {Object} [options] read options
     * @return {Array} an array of the OSC arguments that were read
     */
    osc.readArguments = function (dv, options, offsetState) {
        var typeTagString = osc.readString(dv, offsetState);
        if (typeTagString.indexOf(",") !== 0) {
            // Despite what the OSC 1.0 spec says,
            // it just doesn't make sense to handle messages without type tags.
            // scsynth appears to read such messages as if they have a single
            // Uint8 argument. sclang throws an error if the type tag is omitted.
            throw new Error("A malformed type tag string was found while reading " +
                "the arguments of an OSC message. String was: " +
                typeTagString, " at offset: " + offsetState.idx);
        }

        var argTypes = typeTagString.substring(1).split(""),
            args = [];

        osc.readArgumentsIntoArray(args, argTypes, typeTagString, dv, options, offsetState);

        return args;
    };

    // Unsupported, non-API function.
    osc.readArgument = function (argType, typeTagString, dv, options, offsetState) {
        var typeSpec = osc.argumentTypes[argType];
        if (!typeSpec) {
            throw new Error("'" + argType + "' is not a valid OSC type tag. Type tag string was: " + typeTagString);
        }

        var argReader = typeSpec.reader,
            arg = osc[argReader](dv, offsetState);

        if (options.metadata) {
            arg = {
                type: argType,
                value: arg
            };
        }

        return arg;
    };

    // Unsupported, non-API function.
    osc.readArgumentsIntoArray = function (arr, argTypes, typeTagString, dv, options, offsetState) {
        var i = 0;

        while (i < argTypes.length) {
            var argType = argTypes[i],
                arg;

            if (argType === "[") {
                var fromArrayOpen = argTypes.slice(i + 1),
                    endArrayIdx = fromArrayOpen.indexOf("]");

                if (endArrayIdx < 0) {
                    throw new Error("Invalid argument type tag: an open array type tag ('[') was found " +
                        "without a matching close array tag ('[]'). Type tag was: " + typeTagString);
                }

                var typesInArray = fromArrayOpen.slice(0, endArrayIdx);
                arg = osc.readArgumentsIntoArray([], typesInArray, typeTagString, dv, options, offsetState);
                i += endArrayIdx + 2;
            } else {
                arg = osc.readArgument(argType, typeTagString, dv, options, offsetState);
                i++;
            }

            arr.push(arg);
        }

        return arr;
    };

    /**
     * Writes the specified arguments.
     *
     * @param {Array} args an array of arguments
     * @param {Object} options options for writing
     * @return {Uint8Array} a buffer containing the OSC-formatted argument type tag and values
     */
    osc.writeArguments = function (args, options) {
        var argCollection = osc.collectArguments(args, options);
        return osc.joinParts(argCollection);
    };

    // Unsupported, non-API function.
    osc.joinParts = function (dataCollection) {
        var buf = new Uint8Array(dataCollection.byteLength),
            parts = dataCollection.parts,
            offset = 0;

        for (var i = 0; i < parts.length; i++) {
            var part = parts[i];
            osc.copyByteArray(part, buf, offset);
            offset += part.length;
        }

        return buf;
    };

    // Unsupported, non-API function.
    osc.addDataPart = function (dataPart, dataCollection) {
        dataCollection.parts.push(dataPart);
        dataCollection.byteLength += dataPart.length;
    };

    osc.writeArrayArguments = function (args, dataCollection) {
        var typeTag = "[";

        for (var i = 0; i < args.length; i++) {
            var arg = args[i];
            typeTag += osc.writeArgument(arg, dataCollection);
        }

        typeTag += "]";

        return typeTag;
    };

    osc.writeArgument = function (arg, dataCollection) {
        if (osc.isArray(arg)) {
            return osc.writeArrayArguments(arg, dataCollection);
        }

        var type = arg.type,
            writer = osc.argumentTypes[type].writer;

        if (writer) {
            var data = osc[writer](arg.value);
            osc.addDataPart(data, dataCollection);
        }

        return arg.type;
    };

    // Unsupported, non-API function.
    osc.collectArguments = function (args, options, dataCollection) {
        if (!osc.isArray(args)) {
            args = typeof args === "undefined" ? [] : [args];
        }

        dataCollection = dataCollection || {
            byteLength: 0,
            parts: []
        };

        if (!options.metadata) {
            args = osc.annotateArguments(args);
        }

        var typeTagString = ",",
            currPartIdx = dataCollection.parts.length;

        for (var i = 0; i < args.length; i++) {
            var arg = args[i];
            typeTagString += osc.writeArgument(arg, dataCollection);
        }

        var typeData = osc.writeString(typeTagString);
        dataCollection.byteLength += typeData.byteLength;
        dataCollection.parts.splice(currPartIdx, 0, typeData);

        return dataCollection;
    };

    /**
     * Reads an OSC message.
     *
     * @param {Array-like} data an array of bytes to read from
     * @param {Object} [options] read options
     * @param {Object} [offsetState] an offsetState object that stores the current offset into dv
     * @return {Object} the OSC message, formatted as a JavaScript object containing "address" and "args" properties
     */
    osc.readMessage = function (data, options, offsetState) {
        options = options || osc.defaults;

        var dv = osc.dataView(data, data.byteOffset, data.byteLength);
        offsetState = offsetState || {
            idx: 0
        };

        var address = osc.readString(dv, offsetState);
        return osc.readMessageContents(address, dv, options, offsetState);
    };

    // Unsupported, non-API function.
    osc.readMessageContents = function (address, dv, options, offsetState) {
        if (address.indexOf("/") !== 0) {
            throw new Error("A malformed OSC address was found while reading " +
                "an OSC message. String was: " + address);
        }

        var args = osc.readArguments(dv, options, offsetState);

        return {
            address: address,
            args: args.length === 1 && options.unpackSingleArgs ? args[0] : args
        };
    };

    // Unsupported, non-API function.
    osc.collectMessageParts = function (msg, options, dataCollection) {
        dataCollection = dataCollection || {
            byteLength: 0,
            parts: []
        };

        osc.addDataPart(osc.writeString(msg.address), dataCollection);
        return osc.collectArguments(msg.args, options, dataCollection);
    };

    /**
     * Writes an OSC message.
     *
     * @param {Object} msg a message object containing "address" and "args" properties
     * @param {Object} [options] write options
     * @return {Uint8Array} an array of bytes containing the OSC message
     */
    osc.writeMessage = function (msg, options) {
        options = options || osc.defaults;

        if (!osc.isValidMessage(msg)) {
            throw new Error("An OSC message must contain a valid address. Message was: " +
                JSON.stringify(msg, null, 2));
        }

        var msgCollection = osc.collectMessageParts(msg, options);
        return osc.joinParts(msgCollection);
    };

    osc.isValidMessage = function (msg) {
        return msg.address && msg.address.indexOf("/") === 0;
    };

    /**
     * Reads an OSC bundle.
     *
     * @param {DataView} dv the DataView instance to read from
     * @param {Object} [options] read optoins
     * @param {Object} [offsetState] an offsetState object that stores the current offset into dv
     * @return {Object} the bundle or message object that was read
     */
    osc.readBundle = function (dv, options, offsetState) {
        return osc.readPacket(dv, options, offsetState);
    };

    // Unsupported, non-API function.
    osc.collectBundlePackets = function (bundle, options, dataCollection) {
        dataCollection = dataCollection || {
            byteLength: 0,
            parts: []
        };

        osc.addDataPart(osc.writeString("#bundle"), dataCollection);
        osc.addDataPart(osc.writeTimeTag(bundle.timeTag), dataCollection);

        for (var i = 0; i < bundle.packets.length; i++) {
            var packet = bundle.packets[i],
                collector = packet.address ? osc.collectMessageParts : osc.collectBundlePackets,
                packetCollection = collector(packet, options);

            dataCollection.byteLength += packetCollection.byteLength;
            osc.addDataPart(osc.writeInt32(packetCollection.byteLength), dataCollection);
            dataCollection.parts = dataCollection.parts.concat(packetCollection.parts);
        }

        return dataCollection;
    };

    /**
     * Writes an OSC bundle.
     *
     * @param {Object} a bundle object containing "timeTag" and "packets" properties
     * @param {object} [options] write options
     * @return {Uint8Array} an array of bytes containing the message
     */
    osc.writeBundle = function (bundle, options) {
        if (!osc.isValidBundle(bundle)) {
            throw new Error("An OSC bundle must contain 'timeTag' and 'packets' properties. " +
                "Bundle was: " + JSON.stringify(bundle, null, 2));
        }

        options = options || osc.defaults;
        var bundleCollection = osc.collectBundlePackets(bundle, options);

        return osc.joinParts(bundleCollection);
    };

    osc.isValidBundle = function (bundle) {
        return bundle.timeTag !== undefined && bundle.packets !== undefined;
    };

    // Unsupported, non-API function.
    osc.readBundleContents = function (dv, options, offsetState, len) {
        var timeTag = osc.readTimeTag(dv, offsetState),
            packets = [];

        while (offsetState.idx < len) {
            var packetSize = osc.readInt32(dv, offsetState),
                packetLen = offsetState.idx + packetSize,
                packet = osc.readPacket(dv, options, offsetState, packetLen);

            packets.push(packet);
        }

        return {
            timeTag: timeTag,
            packets: packets
        };
    };

    /**
     * Reads an OSC packet, which may consist of either a bundle or a message.
     *
     * @param {Array-like} data an array of bytes to read from
     * @param {Object} [options] read options
     * @return {Object} a bundle or message object
     */
    osc.readPacket = function (data, options, offsetState, len) {
        var dv = osc.dataView(data, data.byteOffset, data.byteLength);

        len = len === undefined ? dv.byteLength : len;
        offsetState = offsetState || {
            idx: 0
        };

        var header = osc.readString(dv, offsetState),
            firstChar = header[0];

        if (firstChar === "#") {
            return osc.readBundleContents(dv, options, offsetState, len);
        } else if (firstChar === "/") {
            return osc.readMessageContents(header, dv, options, offsetState);
        }

        throw new Error("The header of an OSC packet didn't contain an OSC address or a #bundle string." +
            " Header was: " + header);
    };

    /**
     * Writes an OSC packet, which may consist of either of a bundle or a message.
     *
     * @param {Object} a bundle or message object
     * @param {Object} [options] write options
     * @return {Uint8Array} an array of bytes containing the message
     */
    osc.writePacket = function (packet, options) {
        if (osc.isValidMessage(packet)) {
            return osc.writeMessage(packet, options);
        } else if (osc.isValidBundle(packet)) {
            return osc.writeBundle(packet, options);
        } else {
            throw new Error("The specified packet was not recognized as a valid OSC message or bundle." +
                " Packet was: " + JSON.stringify(packet, null, 2));
        }
    };

    // Unsupported, non-API.
    osc.argumentTypes = {
        i: {
            reader: "readInt32",
            writer: "writeInt32"
        },
        h: {
            reader: "readInt64",
            writer: "writeInt64"
        },
        f: {
            reader: "readFloat32",
            writer: "writeFloat32"
        },
        s: {
            reader: "readString",
            writer: "writeString"
        },
        S: {
            reader: "readString",
            writer: "writeString"
        },
        b: {
            reader: "readBlob",
            writer: "writeBlob"
        },
        t: {
            reader: "readTimeTag",
            writer: "writeTimeTag"
        },
        T: {
            reader: "readTrue"
        },
        F: {
            reader: "readFalse"
        },
        N: {
            reader: "readNull"
        },
        I: {
            reader: "readImpulse"
        },
        d: {
            reader: "readFloat64",
            writer: "writeFloat64"
        },
        c: {
            reader: "readChar32",
            writer: "writeChar32"
        },
        r: {
            reader: "readColor",
            writer: "writeColor"
        },
        m: {
            reader: "readMIDIBytes",
            writer: "writeMIDIBytes"
        },
        // [] are special cased within read/writeArguments()
    };

    // Unsupported, non-API function.
    osc.inferTypeForArgument = function (arg) {
        var type = typeof arg;

        // TODO: This is freaking hideous.
        switch (type) {
            case "boolean":
                return arg ? "T" : "F";
            case "string":
                return "s";
            case "number":
                return "f";
            case "undefined":
                return "N";
            case "object":
                if (arg === null) {
                    return "N";
                } else if (arg instanceof Uint8Array ||
                    arg instanceof ArrayBuffer) {
                    return "b";
                } else if (typeof arg.high === "number" && typeof arg.low === "number") {
                    return "h";
                }
                break;
        }

        throw new Error("Can't infer OSC argument type for value: " +
            JSON.stringify(arg, null, 2));
    };

    // Unsupported, non-API function.
    osc.annotateArguments = function (args) {
        var annotated = [];

        for (var i = 0; i < args.length; i++) {
            var arg = args[i],
                msgArg;

            if (typeof (arg) === "object" && arg.type && arg.value !== undefined) {
                // We've got an explicitly typed argument.
                msgArg = arg;
            } else if (osc.isArray(arg)) {
                // We've got an array of arguments,
                // so they each need to be inferred and expanded.
                msgArg = osc.annotateArguments(arg);
            } else {
                var oscType = osc.inferTypeForArgument(arg);
                msgArg = {
                    type: oscType,
                    value: arg
                };
            }

            annotated.push(msgArg);
        }

        return annotated;
    };

    if (osc.isCommonJS) {
        module.exports = osc;
    }
}());
;
!function(t,i){"object"==typeof exports&&"object"==typeof module?module.exports=i():"function"==typeof define&&define.amd?define([],i):"object"==typeof exports?exports.Long=i():t.Long=i()}("undefined"!=typeof self?self:this,function(){return function(t){function i(e){if(n[e])return n[e].exports;var r=n[e]={i:e,l:!1,exports:{}};return t[e].call(r.exports,r,r.exports,i),r.l=!0,r.exports}var n={};return i.m=t,i.c=n,i.d=function(t,n,e){i.o(t,n)||Object.defineProperty(t,n,{configurable:!1,enumerable:!0,get:e})},i.n=function(t){var n=t&&t.__esModule?function(){return t.default}:function(){return t};return i.d(n,"a",n),n},i.o=function(t,i){return Object.prototype.hasOwnProperty.call(t,i)},i.p="",i(i.s=0)}([function(t,i){function n(t,i,n){this.low=0|t,this.high=0|i,this.unsigned=!!n}function e(t){return!0===(t&&t.__isLong__)}function r(t,i){var n,e,r;return i?(t>>>=0,(r=0<=t&&t<256)&&(e=l[t])?e:(n=h(t,(0|t)<0?-1:0,!0),r&&(l[t]=n),n)):(t|=0,(r=-128<=t&&t<128)&&(e=f[t])?e:(n=h(t,t<0?-1:0,!1),r&&(f[t]=n),n))}function s(t,i){if(isNaN(t))return i?p:m;if(i){if(t<0)return p;if(t>=c)return q}else{if(t<=-v)return _;if(t+1>=v)return E}return t<0?s(-t,i).neg():h(t%d|0,t/d|0,i)}function h(t,i,e){return new n(t,i,e)}function u(t,i,n){if(0===t.length)throw Error("empty string");if("NaN"===t||"Infinity"===t||"+Infinity"===t||"-Infinity"===t)return m;if("number"==typeof i?(n=i,i=!1):i=!!i,(n=n||10)<2||36<n)throw RangeError("radix");var e;if((e=t.indexOf("-"))>0)throw Error("interior hyphen");if(0===e)return u(t.substring(1),i,n).neg();for(var r=s(a(n,8)),h=m,o=0;o<t.length;o+=8){var g=Math.min(8,t.length-o),f=parseInt(t.substring(o,o+g),n);if(g<8){var l=s(a(n,g));h=h.mul(l).add(s(f))}else h=h.mul(r),h=h.add(s(f))}return h.unsigned=i,h}function o(t,i){return"number"==typeof t?s(t,i):"string"==typeof t?u(t,i):h(t.low,t.high,"boolean"==typeof i?i:t.unsigned)}t.exports=n;var g=null;try{g=new WebAssembly.Instance(new WebAssembly.Module(new Uint8Array([0,97,115,109,1,0,0,0,1,13,2,96,0,1,127,96,4,127,127,127,127,1,127,3,7,6,0,1,1,1,1,1,6,6,1,127,1,65,0,11,7,50,6,3,109,117,108,0,1,5,100,105,118,95,115,0,2,5,100,105,118,95,117,0,3,5,114,101,109,95,115,0,4,5,114,101,109,95,117,0,5,8,103,101,116,95,104,105,103,104,0,0,10,191,1,6,4,0,35,0,11,36,1,1,126,32,0,173,32,1,173,66,32,134,132,32,2,173,32,3,173,66,32,134,132,126,34,4,66,32,135,167,36,0,32,4,167,11,36,1,1,126,32,0,173,32,1,173,66,32,134,132,32,2,173,32,3,173,66,32,134,132,127,34,4,66,32,135,167,36,0,32,4,167,11,36,1,1,126,32,0,173,32,1,173,66,32,134,132,32,2,173,32,3,173,66,32,134,132,128,34,4,66,32,135,167,36,0,32,4,167,11,36,1,1,126,32,0,173,32,1,173,66,32,134,132,32,2,173,32,3,173,66,32,134,132,129,34,4,66,32,135,167,36,0,32,4,167,11,36,1,1,126,32,0,173,32,1,173,66,32,134,132,32,2,173,32,3,173,66,32,134,132,130,34,4,66,32,135,167,36,0,32,4,167,11])),{}).exports}catch(t){}n.prototype.__isLong__,Object.defineProperty(n.prototype,"__isLong__",{value:!0}),n.isLong=e;var f={},l={};n.fromInt=r,n.fromNumber=s,n.fromBits=h;var a=Math.pow;n.fromString=u,n.fromValue=o;var d=4294967296,c=d*d,v=c/2,w=r(1<<24),m=r(0);n.ZERO=m;var p=r(0,!0);n.UZERO=p;var y=r(1);n.ONE=y;var b=r(1,!0);n.UONE=b;var N=r(-1);n.NEG_ONE=N;var E=h(-1,2147483647,!1);n.MAX_VALUE=E;var q=h(-1,-1,!0);n.MAX_UNSIGNED_VALUE=q;var _=h(0,-2147483648,!1);n.MIN_VALUE=_;var B=n.prototype;B.toInt=function(){return this.unsigned?this.low>>>0:this.low},B.toNumber=function(){return this.unsigned?(this.high>>>0)*d+(this.low>>>0):this.high*d+(this.low>>>0)},B.toString=function(t){if((t=t||10)<2||36<t)throw RangeError("radix");if(this.isZero())return"0";if(this.isNegative()){if(this.eq(_)){var i=s(t),n=this.div(i),e=n.mul(i).sub(this);return n.toString(t)+e.toInt().toString(t)}return"-"+this.neg().toString(t)}for(var r=s(a(t,6),this.unsigned),h=this,u="";;){var o=h.div(r),g=h.sub(o.mul(r)).toInt()>>>0,f=g.toString(t);if(h=o,h.isZero())return f+u;for(;f.length<6;)f="0"+f;u=""+f+u}},B.getHighBits=function(){return this.high},B.getHighBitsUnsigned=function(){return this.high>>>0},B.getLowBits=function(){return this.low},B.getLowBitsUnsigned=function(){return this.low>>>0},B.getNumBitsAbs=function(){if(this.isNegative())return this.eq(_)?64:this.neg().getNumBitsAbs();for(var t=0!=this.high?this.high:this.low,i=31;i>0&&0==(t&1<<i);i--);return 0!=this.high?i+33:i+1},B.isZero=function(){return 0===this.high&&0===this.low},B.eqz=B.isZero,B.isNegative=function(){return!this.unsigned&&this.high<0},B.isPositive=function(){return this.unsigned||this.high>=0},B.isOdd=function(){return 1==(1&this.low)},B.isEven=function(){return 0==(1&this.low)},B.equals=function(t){return e(t)||(t=o(t)),(this.unsigned===t.unsigned||this.high>>>31!=1||t.high>>>31!=1)&&(this.high===t.high&&this.low===t.low)},B.eq=B.equals,B.notEquals=function(t){return!this.eq(t)},B.neq=B.notEquals,B.ne=B.notEquals,B.lessThan=function(t){return this.comp(t)<0},B.lt=B.lessThan,B.lessThanOrEqual=function(t){return this.comp(t)<=0},B.lte=B.lessThanOrEqual,B.le=B.lessThanOrEqual,B.greaterThan=function(t){return this.comp(t)>0},B.gt=B.greaterThan,B.greaterThanOrEqual=function(t){return this.comp(t)>=0},B.gte=B.greaterThanOrEqual,B.ge=B.greaterThanOrEqual,B.compare=function(t){if(e(t)||(t=o(t)),this.eq(t))return 0;var i=this.isNegative(),n=t.isNegative();return i&&!n?-1:!i&&n?1:this.unsigned?t.high>>>0>this.high>>>0||t.high===this.high&&t.low>>>0>this.low>>>0?-1:1:this.sub(t).isNegative()?-1:1},B.comp=B.compare,B.negate=function(){return!this.unsigned&&this.eq(_)?_:this.not().add(y)},B.neg=B.negate,B.add=function(t){e(t)||(t=o(t));var i=this.high>>>16,n=65535&this.high,r=this.low>>>16,s=65535&this.low,u=t.high>>>16,g=65535&t.high,f=t.low>>>16,l=65535&t.low,a=0,d=0,c=0,v=0;return v+=s+l,c+=v>>>16,v&=65535,c+=r+f,d+=c>>>16,c&=65535,d+=n+g,a+=d>>>16,d&=65535,a+=i+u,a&=65535,h(c<<16|v,a<<16|d,this.unsigned)},B.subtract=function(t){return e(t)||(t=o(t)),this.add(t.neg())},B.sub=B.subtract,B.multiply=function(t){if(this.isZero())return m;if(e(t)||(t=o(t)),g){return h(g.mul(this.low,this.high,t.low,t.high),g.get_high(),this.unsigned)}if(t.isZero())return m;if(this.eq(_))return t.isOdd()?_:m;if(t.eq(_))return this.isOdd()?_:m;if(this.isNegative())return t.isNegative()?this.neg().mul(t.neg()):this.neg().mul(t).neg();if(t.isNegative())return this.mul(t.neg()).neg();if(this.lt(w)&&t.lt(w))return s(this.toNumber()*t.toNumber(),this.unsigned);var i=this.high>>>16,n=65535&this.high,r=this.low>>>16,u=65535&this.low,f=t.high>>>16,l=65535&t.high,a=t.low>>>16,d=65535&t.low,c=0,v=0,p=0,y=0;return y+=u*d,p+=y>>>16,y&=65535,p+=r*d,v+=p>>>16,p&=65535,p+=u*a,v+=p>>>16,p&=65535,v+=n*d,c+=v>>>16,v&=65535,v+=r*a,c+=v>>>16,v&=65535,v+=u*l,c+=v>>>16,v&=65535,c+=i*d+n*a+r*l+u*f,c&=65535,h(p<<16|y,c<<16|v,this.unsigned)},B.mul=B.multiply,B.divide=function(t){if(e(t)||(t=o(t)),t.isZero())throw Error("division by zero");if(g){if(!this.unsigned&&-2147483648===this.high&&-1===t.low&&-1===t.high)return this;return h((this.unsigned?g.div_u:g.div_s)(this.low,this.high,t.low,t.high),g.get_high(),this.unsigned)}if(this.isZero())return this.unsigned?p:m;var i,n,r;if(this.unsigned){if(t.unsigned||(t=t.toUnsigned()),t.gt(this))return p;if(t.gt(this.shru(1)))return b;r=p}else{if(this.eq(_)){if(t.eq(y)||t.eq(N))return _;if(t.eq(_))return y;return i=this.shr(1).div(t).shl(1),i.eq(m)?t.isNegative()?y:N:(n=this.sub(t.mul(i)),r=i.add(n.div(t)))}if(t.eq(_))return this.unsigned?p:m;if(this.isNegative())return t.isNegative()?this.neg().div(t.neg()):this.neg().div(t).neg();if(t.isNegative())return this.div(t.neg()).neg();r=m}for(n=this;n.gte(t);){i=Math.max(1,Math.floor(n.toNumber()/t.toNumber()));for(var u=Math.ceil(Math.log(i)/Math.LN2),f=u<=48?1:a(2,u-48),l=s(i),d=l.mul(t);d.isNegative()||d.gt(n);)i-=f,l=s(i,this.unsigned),d=l.mul(t);l.isZero()&&(l=y),r=r.add(l),n=n.sub(d)}return r},B.div=B.divide,B.modulo=function(t){if(e(t)||(t=o(t)),g){return h((this.unsigned?g.rem_u:g.rem_s)(this.low,this.high,t.low,t.high),g.get_high(),this.unsigned)}return this.sub(this.div(t).mul(t))},B.mod=B.modulo,B.rem=B.modulo,B.not=function(){return h(~this.low,~this.high,this.unsigned)},B.and=function(t){return e(t)||(t=o(t)),h(this.low&t.low,this.high&t.high,this.unsigned)},B.or=function(t){return e(t)||(t=o(t)),h(this.low|t.low,this.high|t.high,this.unsigned)},B.xor=function(t){return e(t)||(t=o(t)),h(this.low^t.low,this.high^t.high,this.unsigned)},B.shiftLeft=function(t){return e(t)&&(t=t.toInt()),0==(t&=63)?this:t<32?h(this.low<<t,this.high<<t|this.low>>>32-t,this.unsigned):h(0,this.low<<t-32,this.unsigned)},B.shl=B.shiftLeft,B.shiftRight=function(t){return e(t)&&(t=t.toInt()),0==(t&=63)?this:t<32?h(this.low>>>t|this.high<<32-t,this.high>>t,this.unsigned):h(this.high>>t-32,this.high>=0?0:-1,this.unsigned)},B.shr=B.shiftRight,B.shiftRightUnsigned=function(t){if(e(t)&&(t=t.toInt()),0===(t&=63))return this;var i=this.high;if(t<32){return h(this.low>>>t|i<<32-t,i>>>t,this.unsigned)}return 32===t?h(i,0,this.unsigned):h(i>>>t-32,0,this.unsigned)},B.shru=B.shiftRightUnsigned,B.shr_u=B.shiftRightUnsigned,B.toSigned=function(){return this.unsigned?h(this.low,this.high,!1):this},B.toUnsigned=function(){return this.unsigned?this:h(this.low,this.high,!0)},B.toBytes=function(t){return t?this.toBytesLE():this.toBytesBE()},B.toBytesLE=function(){var t=this.high,i=this.low;return[255&i,i>>>8&255,i>>>16&255,i>>>24,255&t,t>>>8&255,t>>>16&255,t>>>24]},B.toBytesBE=function(){var t=this.high,i=this.low;return[t>>>24,t>>>16&255,t>>>8&255,255&t,i>>>24,i>>>16&255,i>>>8&255,255&i]},n.fromBytes=function(t,i,e){return e?n.fromBytesLE(t,i):n.fromBytesBE(t,i)},n.fromBytesLE=function(t,i){return new n(t[0]|t[1]<<8|t[2]<<16|t[3]<<24,t[4]|t[5]<<8|t[6]<<16|t[7]<<24,i)},n.fromBytesBE=function(t,i){return new n(t[4]<<24|t[5]<<16|t[6]<<8|t[7],t[0]<<24|t[1]<<16|t[2]<<8|t[3],i)}}])});
/*
 * slip.js: A plain JavaScript SLIP implementation that works in both the browser and Node.js
 *
 * Copyright 2014, Colin Clark
 * Licensed under the MIT and GPL 3 licenses.
 */

/*global exports, define*/
(function (root, factory) {
    "use strict";

    if (typeof exports === "object") {
        // We're in a CommonJS-style loader.
        root.slip = exports;
        factory(exports);
    } else if (typeof define === "function" && define.amd) {
        // We're in an AMD-style loader.
        define(["exports"], function (exports) {
            root.slip = exports;
            return (root.slip, factory(exports));
        });
    } else {
        // Plain old browser.
        root.slip = {};
        factory(root.slip);
    }
}(this, function (exports) {

    "use strict";

    var slip = exports;

    slip.END = 192;
    slip.ESC = 219;
    slip.ESC_END = 220;
    slip.ESC_ESC = 221;

    slip.byteArray = function (data, offset, length) {
        return data instanceof ArrayBuffer ? new Uint8Array(data, offset, length) : data;
    };

    slip.expandByteArray = function (arr) {
        var expanded = new Uint8Array(arr.length * 2);
        expanded.set(arr);

        return expanded;
    };

    slip.sliceByteArray = function (arr, start, end) {
        var sliced = arr.buffer.slice ? arr.buffer.slice(start, end) : arr.subarray(start, end);
        return new Uint8Array(sliced);
    };

    /**
     * SLIP encodes a byte array.
     *
     * @param {Array-like} data a Uint8Array, Node.js Buffer, ArrayBuffer, or [] containing raw bytes
     * @param {Object} options encoder options
     * @return {Uint8Array} the encoded copy of the data
     */
    slip.encode = function (data, o) {
        o = o || {};
        o.bufferPadding = o.bufferPadding || 4; // Will be rounded to the nearest 4 bytes.
        data = slip.byteArray(data, o.offset, o.byteLength);

        var bufLen = (data.length + o.bufferPadding + 3) & ~0x03,
            encoded = new Uint8Array(bufLen),
            j = 1;

        encoded[0] = slip.END;

        for (var i = 0; i < data.length; i++) {
            // We always need enough space for two value bytes plus a trailing END.
            if (j > encoded.length - 3) {
                encoded = slip.expandByteArray(encoded);
            }

            var val = data[i];
            if (val === slip.END) {
                encoded[j++] = slip.ESC;
                val = slip.ESC_END;
            } else if (val === slip.ESC) {
                encoded[j++] = slip.ESC;
                val = slip.ESC_ESC;
            }

            encoded[j++] = val;
        }

        encoded[j] = slip.END;
        return slip.sliceByteArray(encoded, 0, j + 1);
    };

    /**
     * Creates a new SLIP Decoder.
     * @constructor
     *
     * @param {Function} onMessage a callback function that will be invoked when a message has been fully decoded
     * @param {Number} maxBufferSize the maximum size of a incoming message; larger messages will throw an error
     */
    slip.Decoder = function (o) {
        o = typeof o !== "function" ? o || {} : {
            onMessage: o
        };

        this.maxMessageSize = o.maxMessageSize || 10485760; // Defaults to 10 MB.
        this.bufferSize = o.bufferSize || 1024; // Message buffer defaults to 1 KB.
        this.msgBuffer = new Uint8Array(this.bufferSize);
        this.msgBufferIdx = 0;
        this.onMessage = o.onMessage;
        this.onError = o.onError;
        this.escape = false;
    };

    var p = slip.Decoder.prototype;

    /**
     * Decodes a SLIP data packet.
     * The onMessage callback will be invoked when a complete message has been decoded.
     *
     * @param {Array-like} data an incoming stream of bytes
     */
    p.decode = function (data) {
        data = slip.byteArray(data);

        var msg;
        for (var i = 0; i < data.length; i++) {
            var val = data[i];

            if (this.escape) {
                if (val === slip.ESC_ESC) {
                    val = slip.ESC;
                } else if (val === slip.ESC_END) {
                    val = slip.END;
                }
            } else {
                if (val === slip.ESC) {
                    this.escape = true;
                    continue;
                }

                if (val === slip.END) {
                    msg = this.handleEnd();
                    continue;
                }
            }

            var more = this.addByte(val);
            if (!more) {
                this.handleMessageMaxError();
            }
        }

        return msg;
    };

    p.handleMessageMaxError = function () {
        if (this.onError) {
            this.onError(this.msgBuffer.subarray(0),
                "The message is too large; the maximum message size is " +
                this.maxMessageSize / 1024 + "KB. Use a larger maxMessageSize if necessary.");
        }

        // Reset everything and carry on.
        this.msgBufferIdx = 0;
        this.escape = false;
    };

    // Unsupported, non-API method.
    p.addByte = function (val) {
        if (this.msgBufferIdx > this.msgBuffer.length - 1) {
            this.msgBuffer = slip.expandByteArray(this.msgBuffer);
        }

        this.msgBuffer[this.msgBufferIdx++] = val;
        this.escape = false;

        return this.msgBuffer.length < this.maxMessageSize;
    };

    // Unsupported, non-API method.
    p.handleEnd = function () {
        if (this.msgBufferIdx === 0) {
            return; // Toss opening END byte and carry on.
        }

        var msg = slip.sliceByteArray(this.msgBuffer, 0, this.msgBufferIdx);
        if (this.onMessage) {
            this.onMessage(msg);
        }

        // Clear our pointer into the message buffer.
        this.msgBufferIdx = 0;

        return msg;
    };

    return slip;
}));
;
/*!
 * EventEmitter v5.2.9 - git.io/ee
 * Unlicense - http://unlicense.org/
 * Oliver Caldwell - https://oli.me.uk/
 * @preserve
 */

;(function (exports) {
    'use strict';

    /**
     * Class for managing events.
     * Can be extended to provide event functionality in other classes.
     *
     * @class EventEmitter Manages event registering and emitting.
     */
    function EventEmitter() {}

    // Shortcuts to improve speed and size
    var proto = EventEmitter.prototype;
    var originalGlobalValue = exports.EventEmitter;

    /**
     * Finds the index of the listener for the event in its storage array.
     *
     * @param {Function[]} listeners Array of listeners to search through.
     * @param {Function} listener Method to look for.
     * @return {Number} Index of the specified listener, -1 if not found
     * @api private
     */
    function indexOfListener(listeners, listener) {
        var i = listeners.length;
        while (i--) {
            if (listeners[i].listener === listener) {
                return i;
            }
        }

        return -1;
    }

    /**
     * Alias a method while keeping the context correct, to allow for overwriting of target method.
     *
     * @param {String} name The name of the target method.
     * @return {Function} The aliased method
     * @api private
     */
    function alias(name) {
        return function aliasClosure() {
            return this[name].apply(this, arguments);
        };
    }

    /**
     * Returns the listener array for the specified event.
     * Will initialise the event object and listener arrays if required.
     * Will return an object if you use a regex search. The object contains keys for each matched event. So /ba[rz]/ might return an object containing bar and baz. But only if you have either defined them with defineEvent or added some listeners to them.
     * Each property in the object response is an array of listener functions.
     *
     * @param {String|RegExp} evt Name of the event to return the listeners from.
     * @return {Function[]|Object} All listener functions for the event.
     */
    proto.getListeners = function getListeners(evt) {
        var events = this._getEvents();
        var response;
        var key;

        // Return a concatenated array of all matching events if
        // the selector is a regular expression.
        if (evt instanceof RegExp) {
            response = {};
            for (key in events) {
                if (events.hasOwnProperty(key) && evt.test(key)) {
                    response[key] = events[key];
                }
            }
        }
        else {
            response = events[evt] || (events[evt] = []);
        }

        return response;
    };

    /**
     * Takes a list of listener objects and flattens it into a list of listener functions.
     *
     * @param {Object[]} listeners Raw listener objects.
     * @return {Function[]} Just the listener functions.
     */
    proto.flattenListeners = function flattenListeners(listeners) {
        var flatListeners = [];
        var i;

        for (i = 0; i < listeners.length; i += 1) {
            flatListeners.push(listeners[i].listener);
        }

        return flatListeners;
    };

    /**
     * Fetches the requested listeners via getListeners but will always return the results inside an object. This is mainly for internal use but others may find it useful.
     *
     * @param {String|RegExp} evt Name of the event to return the listeners from.
     * @return {Object} All listener functions for an event in an object.
     */
    proto.getListenersAsObject = function getListenersAsObject(evt) {
        var listeners = this.getListeners(evt);
        var response;

        if (listeners instanceof Array) {
            response = {};
            response[evt] = listeners;
        }

        return response || listeners;
    };

    function isValidListener (listener) {
        if (typeof listener === 'function' || listener instanceof RegExp) {
            return true
        } else if (listener && typeof listener === 'object') {
            return isValidListener(listener.listener)
        } else {
            return false
        }
    }

    /**
     * Adds a listener function to the specified event.
     * The listener will not be added if it is a duplicate.
     * If the listener returns true then it will be removed after it is called.
     * If you pass a regular expression as the event name then the listener will be added to all events that match it.
     *
     * @param {String|RegExp} evt Name of the event to attach the listener to.
     * @param {Function} listener Method to be called when the event is emitted. If the function returns true then it will be removed after calling.
     * @return {Object} Current instance of EventEmitter for chaining.
     */
    proto.addListener = function addListener(evt, listener) {
        if (!isValidListener(listener)) {
            throw new TypeError('listener must be a function');
        }

        var listeners = this.getListenersAsObject(evt);
        var listenerIsWrapped = typeof listener === 'object';
        var key;

        for (key in listeners) {
            if (listeners.hasOwnProperty(key) && indexOfListener(listeners[key], listener) === -1) {
                listeners[key].push(listenerIsWrapped ? listener : {
                    listener: listener,
                    once: false
                });
            }
        }

        return this;
    };

    /**
     * Alias of addListener
     */
    proto.on = alias('addListener');

    /**
     * Semi-alias of addListener. It will add a listener that will be
     * automatically removed after its first execution.
     *
     * @param {String|RegExp} evt Name of the event to attach the listener to.
     * @param {Function} listener Method to be called when the event is emitted. If the function returns true then it will be removed after calling.
     * @return {Object} Current instance of EventEmitter for chaining.
     */
    proto.addOnceListener = function addOnceListener(evt, listener) {
        return this.addListener(evt, {
            listener: listener,
            once: true
        });
    };

    /**
     * Alias of addOnceListener.
     */
    proto.once = alias('addOnceListener');

    /**
     * Defines an event name. This is required if you want to use a regex to add a listener to multiple events at once. If you don't do this then how do you expect it to know what event to add to? Should it just add to every possible match for a regex? No. That is scary and bad.
     * You need to tell it what event names should be matched by a regex.
     *
     * @param {String} evt Name of the event to create.
     * @return {Object} Current instance of EventEmitter for chaining.
     */
    proto.defineEvent = function defineEvent(evt) {
        this.getListeners(evt);
        return this;
    };

    /**
     * Uses defineEvent to define multiple events.
     *
     * @param {String[]} evts An array of event names to define.
     * @return {Object} Current instance of EventEmitter for chaining.
     */
    proto.defineEvents = function defineEvents(evts) {
        for (var i = 0; i < evts.length; i += 1) {
            this.defineEvent(evts[i]);
        }
        return this;
    };

    /**
     * Removes a listener function from the specified event.
     * When passed a regular expression as the event name, it will remove the listener from all events that match it.
     *
     * @param {String|RegExp} evt Name of the event to remove the listener from.
     * @param {Function} listener Method to remove from the event.
     * @return {Object} Current instance of EventEmitter for chaining.
     */
    proto.removeListener = function removeListener(evt, listener) {
        var listeners = this.getListenersAsObject(evt);
        var index;
        var key;

        for (key in listeners) {
            if (listeners.hasOwnProperty(key)) {
                index = indexOfListener(listeners[key], listener);

                if (index !== -1) {
                    listeners[key].splice(index, 1);
                }
            }
        }

        return this;
    };

    /**
     * Alias of removeListener
     */
    proto.off = alias('removeListener');

    /**
     * Adds listeners in bulk using the manipulateListeners method.
     * If you pass an object as the first argument you can add to multiple events at once. The object should contain key value pairs of events and listeners or listener arrays. You can also pass it an event name and an array of listeners to be added.
     * You can also pass it a regular expression to add the array of listeners to all events that match it.
     * Yeah, this function does quite a bit. That's probably a bad thing.
     *
     * @param {String|Object|RegExp} evt An event name if you will pass an array of listeners next. An object if you wish to add to multiple events at once.
     * @param {Function[]} [listeners] An optional array of listener functions to add.
     * @return {Object} Current instance of EventEmitter for chaining.
     */
    proto.addListeners = function addListeners(evt, listeners) {
        // Pass through to manipulateListeners
        return this.manipulateListeners(false, evt, listeners);
    };

    /**
     * Removes listeners in bulk using the manipulateListeners method.
     * If you pass an object as the first argument you can remove from multiple events at once. The object should contain key value pairs of events and listeners or listener arrays.
     * You can also pass it an event name and an array of listeners to be removed.
     * You can also pass it a regular expression to remove the listeners from all events that match it.
     *
     * @param {String|Object|RegExp} evt An event name if you will pass an array of listeners next. An object if you wish to remove from multiple events at once.
     * @param {Function[]} [listeners] An optional array of listener functions to remove.
     * @return {Object} Current instance of EventEmitter for chaining.
     */
    proto.removeListeners = function removeListeners(evt, listeners) {
        // Pass through to manipulateListeners
        return this.manipulateListeners(true, evt, listeners);
    };

    /**
     * Edits listeners in bulk. The addListeners and removeListeners methods both use this to do their job. You should really use those instead, this is a little lower level.
     * The first argument will determine if the listeners are removed (true) or added (false).
     * If you pass an object as the second argument you can add/remove from multiple events at once. The object should contain key value pairs of events and listeners or listener arrays.
     * You can also pass it an event name and an array of listeners to be added/removed.
     * You can also pass it a regular expression to manipulate the listeners of all events that match it.
     *
     * @param {Boolean} remove True if you want to remove listeners, false if you want to add.
     * @param {String|Object|RegExp} evt An event name if you will pass an array of listeners next. An object if you wish to add/remove from multiple events at once.
     * @param {Function[]} [listeners] An optional array of listener functions to add/remove.
     * @return {Object} Current instance of EventEmitter for chaining.
     */
    proto.manipulateListeners = function manipulateListeners(remove, evt, listeners) {
        var i;
        var value;
        var single = remove ? this.removeListener : this.addListener;
        var multiple = remove ? this.removeListeners : this.addListeners;

        // If evt is an object then pass each of its properties to this method
        if (typeof evt === 'object' && !(evt instanceof RegExp)) {
            for (i in evt) {
                if (evt.hasOwnProperty(i) && (value = evt[i])) {
                    // Pass the single listener straight through to the singular method
                    if (typeof value === 'function') {
                        single.call(this, i, value);
                    }
                    else {
                        // Otherwise pass back to the multiple function
                        multiple.call(this, i, value);
                    }
                }
            }
        }
        else {
            // So evt must be a string
            // And listeners must be an array of listeners
            // Loop over it and pass each one to the multiple method
            i = listeners.length;
            while (i--) {
                single.call(this, evt, listeners[i]);
            }
        }

        return this;
    };

    /**
     * Removes all listeners from a specified event.
     * If you do not specify an event then all listeners will be removed.
     * That means every event will be emptied.
     * You can also pass a regex to remove all events that match it.
     *
     * @param {String|RegExp} [evt] Optional name of the event to remove all listeners for. Will remove from every event if not passed.
     * @return {Object} Current instance of EventEmitter for chaining.
     */
    proto.removeEvent = function removeEvent(evt) {
        var type = typeof evt;
        var events = this._getEvents();
        var key;

        // Remove different things depending on the state of evt
        if (type === 'string') {
            // Remove all listeners for the specified event
            delete events[evt];
        }
        else if (evt instanceof RegExp) {
            // Remove all events matching the regex.
            for (key in events) {
                if (events.hasOwnProperty(key) && evt.test(key)) {
                    delete events[key];
                }
            }
        }
        else {
            // Remove all listeners in all events
            delete this._events;
        }

        return this;
    };

    /**
     * Alias of removeEvent.
     *
     * Added to mirror the node API.
     */
    proto.removeAllListeners = alias('removeEvent');

    /**
     * Emits an event of your choice.
     * When emitted, every listener attached to that event will be executed.
     * If you pass the optional argument array then those arguments will be passed to every listener upon execution.
     * Because it uses `apply`, your array of arguments will be passed as if you wrote them out separately.
     * So they will not arrive within the array on the other side, they will be separate.
     * You can also pass a regular expression to emit to all events that match it.
     *
     * @param {String|RegExp} evt Name of the event to emit and execute listeners for.
     * @param {Array} [args] Optional array of arguments to be passed to each listener.
     * @return {Object} Current instance of EventEmitter for chaining.
     */
    proto.emitEvent = function emitEvent(evt, args) {
        var listenersMap = this.getListenersAsObject(evt);
        var listeners;
        var listener;
        var i;
        var key;
        var response;

        for (key in listenersMap) {
            if (listenersMap.hasOwnProperty(key)) {
                listeners = listenersMap[key].slice(0);

                for (i = 0; i < listeners.length; i++) {
                    // If the listener returns true then it shall be removed from the event
                    // The function is executed either with a basic call or an apply if there is an args array
                    listener = listeners[i];

                    if (listener.once === true) {
                        this.removeListener(evt, listener.listener);
                    }

                    response = listener.listener.apply(this, args || []);

                    if (response === this._getOnceReturnValue()) {
                        this.removeListener(evt, listener.listener);
                    }
                }
            }
        }

        return this;
    };

    /**
     * Alias of emitEvent
     */
    proto.trigger = alias('emitEvent');

    /**
     * Subtly different from emitEvent in that it will pass its arguments on to the listeners, as opposed to taking a single array of arguments to pass on.
     * As with emitEvent, you can pass a regex in place of the event name to emit to all events that match it.
     *
     * @param {String|RegExp} evt Name of the event to emit and execute listeners for.
     * @param {...*} Optional additional arguments to be passed to each listener.
     * @return {Object} Current instance of EventEmitter for chaining.
     */
    proto.emit = function emit(evt) {
        var args = Array.prototype.slice.call(arguments, 1);
        return this.emitEvent(evt, args);
    };

    /**
     * Sets the current value to check against when executing listeners. If a
     * listeners return value matches the one set here then it will be removed
     * after execution. This value defaults to true.
     *
     * @param {*} value The new value to check for when executing listeners.
     * @return {Object} Current instance of EventEmitter for chaining.
     */
    proto.setOnceReturnValue = function setOnceReturnValue(value) {
        this._onceReturnValue = value;
        return this;
    };

    /**
     * Fetches the current value to check against when executing listeners. If
     * the listeners return value matches this one then it should be removed
     * automatically. It will return true by default.
     *
     * @return {*|Boolean} The current value to check for or the default, true.
     * @api private
     */
    proto._getOnceReturnValue = function _getOnceReturnValue() {
        if (this.hasOwnProperty('_onceReturnValue')) {
            return this._onceReturnValue;
        }
        else {
            return true;
        }
    };

    /**
     * Fetches the events object and creates one if required.
     *
     * @return {Object} The events storage object.
     * @api private
     */
    proto._getEvents = function _getEvents() {
        return this._events || (this._events = {});
    };

    /**
     * Reverts the global {@link EventEmitter} to its previous value and returns a reference to this version.
     *
     * @return {Function} Non conflicting EventEmitter class.
     */
    EventEmitter.noConflict = function noConflict() {
        exports.EventEmitter = originalGlobalValue;
        return EventEmitter;
    };

    // Expose the class either via AMD, CommonJS or the global object
    if (typeof define === 'function' && define.amd) {
        define(function () {
            return EventEmitter;
        });
    }
    else if (typeof module === 'object' && module.exports){
        module.exports = EventEmitter;
    }
    else {
        exports.EventEmitter = EventEmitter;
    }
}(typeof window !== 'undefined' ? window : this || {}));
;
/*
 * osc.js: An Open Sound Control library for JavaScript that works in both the browser and Node.js
 *
 * Cross-platform base transport library for osc.js.
 *
 * Copyright 2014-2016, Colin Clark
 * Licensed under the MIT and GPL 3 licenses.
 */

/* global require, module */

var osc = osc || require("./osc.js"),
    slip = slip || require("slip"),
    EventEmitter = EventEmitter || require("events").EventEmitter;

(function () {

    "use strict";

    osc.supportsSerial = false;

    // Unsupported, non-API function.
    osc.firePacketEvents = function (port, packet, timeTag, packetInfo) {
        if (packet.address) {
            port.emit("message", packet, timeTag, packetInfo);
        } else {
            osc.fireBundleEvents(port, packet, timeTag, packetInfo);
        }
    };

    // Unsupported, non-API function.
    osc.fireBundleEvents = function (port, bundle, timeTag, packetInfo) {
        port.emit("bundle", bundle, timeTag, packetInfo);
        for (var i = 0; i < bundle.packets.length; i++) {
            var packet = bundle.packets[i];
            osc.firePacketEvents(port, packet, bundle.timeTag, packetInfo);
        }
    };

    osc.fireClosedPortSendError = function (port, msg) {
        msg = msg || "Can't send packets on a closed osc.Port object. Please open (or reopen) this Port by calling open().";

        port.emit("error", msg);
    };

    osc.Port = function (options) {
        this.options = options || {};
        this.on("data", this.decodeOSC.bind(this));
    };

    var p = osc.Port.prototype = Object.create(EventEmitter.prototype);
    p.constructor = osc.Port;

    p.send = function (oscPacket) {
        var args = Array.prototype.slice.call(arguments),
            encoded = this.encodeOSC(oscPacket),
            buf = osc.nativeBuffer(encoded);

        args[0] = buf;
        this.sendRaw.apply(this, args);
    };

    p.encodeOSC = function (packet) {
        // TODO gh-39: This is unsafe; we should only access the underlying
        // buffer within the range of its view.
        packet = packet.buffer ? packet.buffer : packet;
        var encoded;

        try {
            encoded = osc.writePacket(packet, this.options);
        } catch (err) {
            this.emit("error", err);
        }

        return encoded;
    };

    p.decodeOSC = function (data, packetInfo) {
        data = osc.byteArray(data);
        this.emit("raw", data, packetInfo);

        try {
            var packet = osc.readPacket(data, this.options);
            this.emit("osc", packet, packetInfo);
            osc.firePacketEvents(this, packet, undefined, packetInfo);
        } catch (err) {
            this.emit("error", err);
        }
    };


    osc.SLIPPort = function (options) {
        var that = this;
        var o = this.options = options || {};
        o.useSLIP = o.useSLIP === undefined ? true : o.useSLIP;

        this.decoder = new slip.Decoder({
            onMessage: this.decodeOSC.bind(this),
            onError: function (err) {
                that.emit("error", err);
            }
        });

        var decodeHandler = o.useSLIP ? this.decodeSLIPData : this.decodeOSC;
        this.on("data", decodeHandler.bind(this));
    };

    p = osc.SLIPPort.prototype = Object.create(osc.Port.prototype);
    p.constructor = osc.SLIPPort;

    p.encodeOSC = function (packet) {
        // TODO gh-39: This is unsafe; we should only access the underlying
        // buffer within the range of its view.
        packet = packet.buffer ? packet.buffer : packet;
        var framed;

        try {
            var encoded = osc.writePacket(packet, this.options);
            framed = slip.encode(encoded);
        } catch (err) {
            this.emit("error", err);
        }

        return framed;
    };

    p.decodeSLIPData = function (data, packetInfo) {
        // TODO: Get packetInfo through SLIP decoder.
        this.decoder.decode(data, packetInfo);
    };


    // Unsupported, non-API function.
    osc.relay = function (from, to, eventName, sendFnName, transformFn, sendArgs) {
        eventName = eventName || "message";
        sendFnName = sendFnName || "send";
        transformFn = transformFn || function () {};
        sendArgs = sendArgs ? [null].concat(sendArgs) : [];

        var listener = function (data) {
            sendArgs[0] = data;
            data = transformFn(data);
            to[sendFnName].apply(to, sendArgs);
        };

        from.on(eventName, listener);

        return {
            eventName: eventName,
            listener: listener
        };
    };

    // Unsupported, non-API function.
    osc.relayPorts = function (from, to, o) {
        var eventName = o.raw ? "raw" : "osc",
            sendFnName = o.raw ? "sendRaw" : "send";

        return osc.relay(from, to, eventName, sendFnName, o.transform);
    };

    // Unsupported, non-API function.
    osc.stopRelaying = function (from, relaySpec) {
        from.removeListener(relaySpec.eventName, relaySpec.listener);
    };


    /**
     * A Relay connects two sources of OSC data together,
     * relaying all OSC messages received by each port to the other.
     * @constructor
     *
     * @param {osc.Port} port1 the first port to relay
     * @param {osc.Port} port2 the second port to relay
     * @param {Object} options the configuration options for this relay
     */
    osc.Relay = function (port1, port2, options) {
        var o = this.options = options || {};
        o.raw = false;

        this.port1 = port1;
        this.port2 = port2;

        this.listen();
    };

    p = osc.Relay.prototype = Object.create(EventEmitter.prototype);
    p.constructor = osc.Relay;

    p.open = function () {
        this.port1.open();
        this.port2.open();
    };

    p.listen = function () {
        if (this.port1Spec && this.port2Spec) {
            this.close();
        }

        this.port1Spec = osc.relayPorts(this.port1, this.port2, this.options);
        this.port2Spec = osc.relayPorts(this.port2, this.port1, this.options);

        // Bind port close listeners to ensure that the relay
        // will stop forwarding messages if one of its ports close.
        // Users are still responsible for closing the underlying ports
        // if necessary.
        var closeListener = this.close.bind(this);
        this.port1.on("close", closeListener);
        this.port2.on("close", closeListener);
    };

    p.close = function () {
        osc.stopRelaying(this.port1, this.port1Spec);
        osc.stopRelaying(this.port2, this.port2Spec);
        this.emit("close", this.port1, this.port2);
    };


    // If we're in a require-compatible environment, export ourselves.
    if (typeof module !== "undefined" && module.exports) {
        module.exports = osc;
    }
}());
;
/*
 * osc.js: An Open Sound Control library for JavaScript that works in both the browser and Node.js
 *
 * Cross-Platform Web Socket client transport for osc.js.
 *
 * Copyright 2014-2016, Colin Clark
 * Licensed under the MIT and GPL 3 licenses.
 */

/*global WebSocket, require*/

var osc = osc || require("./osc.js");

(function () {

    "use strict";

    osc.WebSocket = typeof WebSocket !== "undefined" ? WebSocket : require ("ws");

    osc.WebSocketPort = function (options) {
        osc.Port.call(this, options);
        this.on("open", this.listen.bind(this));

        this.socket = options.socket;
        if (this.socket) {
            if (this.socket.readyState === 1) {
                osc.WebSocketPort.setupSocketForBinary(this.socket);
                this.emit("open", this.socket);
            } else {
                this.open();
            }
        }
    };

    var p = osc.WebSocketPort.prototype = Object.create(osc.Port.prototype);
    p.constructor = osc.WebSocketPort;

    p.open = function () {
        if (!this.socket || this.socket.readyState > 1) {
            this.socket = new osc.WebSocket(this.options.url);
        }

        osc.WebSocketPort.setupSocketForBinary(this.socket);

        var that = this;
        this.socket.onopen = function () {
            that.emit("open", that.socket);
        };

        this.socket.onerror = function (err) {
            that.emit("error", err);
        };
    };

    p.listen = function () {
        var that = this;
        this.socket.onmessage = function (e) {
            that.emit("data", e.data, e);
        };

        this.socket.onclose = function (e) {
            that.emit("close", e);
        };

        that.emit("ready");
    };

    p.sendRaw = function (encoded) {
        if (!this.socket || this.socket.readyState !== 1) {
            osc.fireClosedPortSendError(this);
            return;
        }

        this.socket.send(encoded);
    };

    p.close = function (code, reason) {
        this.socket.close(code, reason);
    };

    osc.WebSocketPort.setupSocketForBinary = function (socket) {
        socket.binaryType = osc.isNode ? "nodebuffer" : "arraybuffer";
    };

}());
