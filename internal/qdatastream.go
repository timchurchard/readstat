package internal

import (
	"encoding/binary"
	"fmt"
	"io"
	"net/url"
	"time"
	"unicode/utf16"
)

// From: https://www.mobileread.com/forums/showthread.php?t=313990
// From: https://go.dev/play/p/EnWAEYdmunk
// From: https://gist.github.com/pgaskin/a41a61ffe6d70567a11dc481020c5290
// Credit to github.com/pgaskin
//
// An implementation of a subset of the Qt 5.2 QDataStream serialization:
//
// https://doc.qt.io/archives/qt-4.8/datastreamformat.html
// https://github.com/kobolabs/qtbase/blob/509a5ff05334da08f906cbc1f719f00a74341fc4/src/corelib/io/qdatastream.cpp
// https://github.com/kobolabs/qtbase/blob/509a5ff05334da08f906cbc1f719f00a74341fc4/src/corelib/io/qdatastream.h
// https://github.com/kobolabs/qtbase/blob/3f3a6f7b8ec0198c8ad993b613a0ef9152280c5c/src/corelib/kernel/qmetatype.h#L73-L154
//
// Example
/*func main() {
	// An EventData blob from the Events table in the Kobo eReader firmware.
	b, err := hex.DecodeString("0000000400000010005600690065007700540079007000650000000a00000000060054004f00430000003000450078007400720061004400610074006100520065006100640069006e006700530065007300730069006f006e00730000000200000000030000002e00450078007400720061004400610074006100520065006100640069006e0067005300650063006f006e0064007300000002000000000a00000028004500780074007200610044006100740061004400610074006500430072006500610074006500640000000a00000000280032003000310039002d00310031002d00320035005400300031003a00310037003a00310034005a")
	if err != nil {
		panic(err)
	}
	r := bytes.NewBuffer(b)
	v, err := (&QDataStreamReader{
		Reader:    r,
		ByteOrder: binary.BigEndian,
	}).ReadQStringQVariantAssociative()
	if err != nil {
		panic(err)
	}
	json.NewEncoder(os.Stdout).Encode(v)
	if r.Len() != 0 {
		panic("not all read")
	}
}*/

// An implementation of a subset of the Qt 5.2 QDataStream serialization of a
// QMap<QString, QVariant>. See:
//
// https://doc.qt.io/archives/qt-4.8/datastreamformat.html
// https://github.com/kobolabs/qtbase/blob/509a5ff05334da08f906cbc1f719f00a74341fc4/src/corelib/io/qdatastream.cpp
// https://github.com/kobolabs/qtbase/blob/509a5ff05334da08f906cbc1f719f00a74341fc4/src/corelib/io/qdatastream.h
// https://github.com/kobolabs/qtbase/blob/3f3a6f7b8ec0198c8ad993b613a0ef9152280c5c/src/corelib/kernel/qmetatype.h#L73-L154

// QMetaType represents a Qt metatype.
type QMetaType int

// From qtbase/corelib/kernel/qmetatype.h.
const (
	QMetaTypeBool               QMetaType = 1
	QMetaTypeInt                QMetaType = 2
	QMetaTypeUInt               QMetaType = 3
	QMetaTypeLongLong           QMetaType = 4
	QMetaTypeULongLong          QMetaType = 5
	QMetaTypeDouble             QMetaType = 6
	QMetaTypeQChar              QMetaType = 7
	QMetaTypeQVariantMap        QMetaType = 8
	QMetaTypeQVariantList       QMetaType = 9
	QMetaTypeQString            QMetaType = 10
	QMetaTypeQStringList        QMetaType = 11
	QMetaTypeQByteArray         QMetaType = 12
	QMetaTypeQBitArray          QMetaType = 13
	QMetaTypeQDate              QMetaType = 14
	QMetaTypeQTime              QMetaType = 15
	QMetaTypeQDateTime          QMetaType = 16
	QMetaTypeQUrl               QMetaType = 17
	QMetaTypeQLocale            QMetaType = 18
	QMetaTypeQRect              QMetaType = 19
	QMetaTypeQRectF             QMetaType = 20
	QMetaTypeQSize              QMetaType = 21
	QMetaTypeQSizeF             QMetaType = 22
	QMetaTypeQLine              QMetaType = 23
	QMetaTypeQLineF             QMetaType = 24
	QMetaTypeQPoint             QMetaType = 25
	QMetaTypeQPointF            QMetaType = 26
	QMetaTypeQRegExp            QMetaType = 27
	QMetaTypeQVariantHash       QMetaType = 28
	QMetaTypeQEasingCurve       QMetaType = 29
	QMetaTypeQUuid              QMetaType = 30
	QMetaTypeVoidStar           QMetaType = 31
	QMetaTypeLong               QMetaType = 32
	QMetaTypeShort              QMetaType = 33
	QMetaTypeChar               QMetaType = 34
	QMetaTypeULong              QMetaType = 35
	QMetaTypeUShort             QMetaType = 36
	QMetaTypeUChar              QMetaType = 37
	QMetaTypeFloat              QMetaType = 38
	QMetaTypeQObjectStar        QMetaType = 39
	QMetaTypeSChar              QMetaType = 40
	QMetaTypeVoid               QMetaType = 43
	QMetaTypeQVariant           QMetaType = 41
	QMetaTypeQModelIndex        QMetaType = 42
	QMetaTypeQRegularExpression QMetaType = 44
	QMetaTypeQJsonValue         QMetaType = 45
	QMetaTypeQJsonObject        QMetaType = 46
	QMetaTypeQJsonArray         QMetaType = 47
	QMetaTypeQJsonDocument      QMetaType = 48
	QMetaTypeQFont              QMetaType = 64
	QMetaTypeQPixmap            QMetaType = 65
	QMetaTypeQBrush             QMetaType = 66
	QMetaTypeQColor             QMetaType = 67
	QMetaTypeQPalette           QMetaType = 68
	QMetaTypeQIcon              QMetaType = 69
	QMetaTypeQImage             QMetaType = 70
	QMetaTypeQPolygon           QMetaType = 71
	QMetaTypeQRegion            QMetaType = 72
	QMetaTypeQBitmap            QMetaType = 73
	QMetaTypeQCursor            QMetaType = 74
	QMetaTypeQKeySequence       QMetaType = 75
	QMetaTypeQPen               QMetaType = 76
	QMetaTypeQTextLength        QMetaType = 77
	QMetaTypeQTextFormat        QMetaType = 78
	QMetaTypeQMatrix            QMetaType = 79
	QMetaTypeQTransform         QMetaType = 80
	QMetaTypeQMatrix4x4         QMetaType = 81
	QMetaTypeQVector2D          QMetaType = 82
	QMetaTypeQVector3D          QMetaType = 83
	QMetaTypeQVector4D          QMetaType = 84
	QMetaTypeQQuaternion        QMetaType = 85
	QMetaTypeQPolygonF          QMetaType = 86
	QMetaTypeQSizePolicy        QMetaType = 121
)

// QDataStreamReader parses a Qt 4.5-5.2 QDataStream (other versions will
// probably work, as this only implements a small subset).
type QDataStreamReader struct {
	Reader    io.Reader
	ByteOrder binary.ByteOrder
}

func (r *QDataStreamReader) ReadBool() (bool, error) {
	var v uint8
	if err := binary.Read(r.Reader, r.ByteOrder, &v); err != nil {
		return false, err
	}
	return v != 0, nil
}

func (r *QDataStreamReader) ReadInt8() (int8, error) {
	var v int8
	if err := binary.Read(r.Reader, r.ByteOrder, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func (r *QDataStreamReader) ReadInt16() (int16, error) {
	var v int16
	if err := binary.Read(r.Reader, r.ByteOrder, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func (r *QDataStreamReader) ReadInt32() (int32, error) {
	var v int32
	if err := binary.Read(r.Reader, r.ByteOrder, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func (r *QDataStreamReader) ReadInt64() (int64, error) {
	var v int64
	if err := binary.Read(r.Reader, r.ByteOrder, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func (r *QDataStreamReader) ReadUint8() (uint8, error) {
	var v uint8
	if err := binary.Read(r.Reader, r.ByteOrder, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func (r *QDataStreamReader) ReadUint16() (uint16, error) {
	var v uint16
	if err := binary.Read(r.Reader, r.ByteOrder, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func (r *QDataStreamReader) ReadUint32() (uint32, error) {
	var v uint32
	if err := binary.Read(r.Reader, r.ByteOrder, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func (r *QDataStreamReader) ReadUint64() (uint64, error) {
	var v uint64
	if err := binary.Read(r.Reader, r.ByteOrder, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func (r *QDataStreamReader) ReadFloat() (float32, error) {
	var v float32
	if err := binary.Read(r.Reader, r.ByteOrder, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func (r *QDataStreamReader) ReadDouble() (float64, error) {
	var v float64
	if err := binary.Read(r.Reader, r.ByteOrder, &v); err != nil {
		return 0, err
	}
	return v, nil
}

func (r *QDataStreamReader) ReadCString() (string, error) {
	n, err := r.ReadUint32()
	if err != nil {
		return "", err
	}
	buf := make([]byte, n)
	if err := binary.Read(r.Reader, r.ByteOrder, &buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func (r *QDataStreamReader) ReadQBitArray() ([]bool, error) {
	n, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}
	buf := make([]byte, (n+7)/8)
	if err := binary.Read(r.Reader, r.ByteOrder, &buf); err != nil {
		return nil, err
	}
	bits := make([]bool, n)
	for i := n - 1; i >= 0; i-- {
		bits[i] = (((buf[i/8]) >> (7 - i%8)) & 0x1) == 0x1
	}
	return bits, nil
}

func (r *QDataStreamReader) ReadQByteArray() ([]byte, error) {
	n, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}
	if n == 0xFFFFFFFF {
		return nil, nil
	}
	buf := make([]byte, n)
	if err := binary.Read(r.Reader, r.ByteOrder, &buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (r *QDataStreamReader) ReadQDate() (time.Time, error) {
	julian, err := r.ReadUint32()
	if err != nil {
		return time.Time{}, err
	}
	// ported from qdatetime.cpp
	floordiv := func(a, b int) int {
		var x int
		if a < 0 {
			x = b - 1
		}
		return (a - x) / b
	}
	var a, b int64
	a = int64(julian) + 32044
	b = int64(floordiv(int(4*a+3), 146097))
	var c, d, e, m, day, month, year int
	c = int(a) - floordiv(146097*int(b), 4)
	d = floordiv(4*c+3, 1461)
	e = c - floordiv(1461*d, 4)
	m = floordiv(5*e+2, 153)
	day = e - floordiv(153*m+2, 5) + 1
	month = m + 3 - 12*floordiv(m, 10)
	year = 100*int(b) + d - 4800 + floordiv(m, 10)
	if year <= 0 {
		year--
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

func (r *QDataStreamReader) ReadQString() (string, error) {
	n, err := r.ReadUint32()
	if err != nil {
		return "", err
	}
	if n == 0xFFFFFFFF {
		return "", nil
	}
	buf := make([]uint16, n/2)
	if err := binary.Read(r.Reader, r.ByteOrder, &buf); err != nil {
		return "", err
	}
	return string(utf16.Decode(buf)), nil
}

func (r *QDataStreamReader) ReadQTime() (time.Duration, error) { // msecs past midnight
	msecsMidnight, err := r.ReadUint32()
	if err != nil {
		return 0, err
	}
	return time.Millisecond * time.Duration(msecsMidnight), nil
}

func (r *QDataStreamReader) ReadQUrl() (*url.URL, error) { // msecs past midnight
	str, err := r.ReadQString()
	if err != nil {
		return nil, err
	}
	if str == "" {
		return nil, nil
	}
	return url.Parse(str)
}

func (r *QDataStreamReader) ReadQVariant() (QMetaType, interface{}, error) { // msecs past midnight
	t, err := r.ReadUint32()
	if err != nil {
		return 0, nil, err
	}
	if null, err := r.ReadBool(); err != nil {
		return 0, nil, err
	} else if null {
		return QMetaType(t), nil, nil
	}

	var v interface{}
	err = nil
	switch QMetaType(t) {
	case QMetaTypeBool:
		x, xerr := r.ReadBool()
		v, err = bool(x), xerr
	case QMetaTypeInt:
		x, xerr := r.ReadInt32()
		v, err = int32(x), xerr
	case QMetaTypeUInt:
		x, xerr := r.ReadUint32()
		v, err = uint32(x), xerr
	case QMetaTypeLongLong:
		x, xerr := r.ReadInt64()
		v, err = int64(x), xerr
	case QMetaTypeULongLong:
		x, xerr := r.ReadUint64()
		v, err = uint64(x), xerr
	case QMetaTypeDouble:
		x, xerr := r.ReadDouble()
		v, err = float64(x), xerr
	case QMetaTypeFloat:
		x, xerr := r.ReadFloat()
		v, err = float32(x), xerr
	case QMetaTypeQChar, QMetaTypeChar:
		x, xerr := r.ReadUint8()
		v, err = uint8(x), xerr
	case QMetaTypeSChar:
		x, xerr := r.ReadInt8()
		v, err = int8(x), xerr
	case QMetaTypeShort:
		x, xerr := r.ReadInt16()
		v, err = int16(x), xerr
	case QMetaTypeUShort:
		x, xerr := r.ReadUint16()
		v, err = uint16(x), xerr
	case QMetaTypeQBitArray:
		x, xerr := r.ReadQBitArray()
		v, err = []bool(x), xerr
	case QMetaTypeQVariantMap, QMetaTypeQVariantHash:
		x, xerr := r.ReadQStringQVariantAssociative()
		v, err = map[string]interface{}(x), xerr
	case QMetaTypeQVariantList:
		x, xerr := r.ReadQStringQVariantList()
		v, err = []interface{}(x), xerr
	case QMetaTypeQByteArray:
		x, xerr := r.ReadQByteArray()
		v, err = []byte(x), xerr
	case QMetaTypeQString:
		x, xerr := r.ReadQString()
		v, err = string(x), xerr
	case QMetaTypeQStringList:
		x, xerr := r.ReadQStringQStringList()
		v, err = []string(x), xerr
	case QMetaTypeQDate:
		x, xerr := r.ReadQDate()
		v, err = time.Time(x), xerr
	case QMetaTypeQTime:
		x, xerr := r.ReadQTime()
		v, err = time.Duration(x), xerr
	case QMetaTypeQDateTime:
		x, xerr := r.ReadQDateTime()
		v, err = time.Time(x), xerr
	case QMetaTypeQUrl:
		x, xerr := r.ReadQUrl()
		v, err = (*url.URL)(x), xerr
	default:
		return QMetaType(t), nil, fmt.Errorf("unimplemented type %d", t)
	}
	return QMetaType(t), v, err
}

func (r *QDataStreamReader) ReadQDateTime() (time.Time, error) {
	d, err := r.ReadQDate()
	if err != nil {
		return time.Time{}, err
	}
	t, err := r.ReadQTime()
	if err != nil {
		return time.Time{}, err
	}
	u, err := r.ReadUint8()
	if err != nil {
		return time.Time{}, err
	}
	var z *time.Location
	if u == 0 {
		z = time.Local
	} else {
		z = time.UTC
	}
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, z).Add(t), nil
}

func (r *QDataStreamReader) ReadQStringQVariantList() ([]interface{}, error) {
	n, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m := make([]interface{}, n)
	for i := range m {
		_, v, err := r.ReadQVariant()
		if err != nil {
			return nil, err
		}
		m[i] = v
	}
	return m, nil
}

func (r *QDataStreamReader) ReadQStringQStringList() ([]string, error) {
	n, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m := make([]string, n)
	for i := range m {
		v, err := r.ReadQString()
		if err != nil {
			return nil, err
		}
		m[i] = v
	}
	return m, nil
}

func (r *QDataStreamReader) ReadQStringQVariantAssociative() (map[string]interface{}, error) {
	n, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m := map[string]interface{}{}
	for i := uint32(0); i < n; i++ {
		k, err := r.ReadQString()
		if err != nil {
			return m, err
		}
		_, v, err := r.ReadQVariant()
		if err != nil {
			return m, err
		}
		m[k] = v
	}
	return m, nil
}
