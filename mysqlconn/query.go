package mysqlconn

import (
	"fmt"

	"github.com/youtube/vitess/go/sqldb"
	"github.com/youtube/vitess/go/sqltypes"

	querypb "github.com/youtube/vitess/go/vt/proto/query"
)

// This file contains the methods related to queries.

func (c *Conn) writeComQuery(query string) error {
	data := make([]byte, len(query)+1)
	data[0] = ComQuery
	copy(data[1:], []byte(query))
	if err := c.writePacket(data); err != nil {
		return err
	}
	if err := c.flush(); err != nil {
		return err
	}
	return nil
}

func (c *Conn) writeComInitDB(db string) error {
	data := make([]byte, len(db)+1)
	data[0] = ComInitDB
	copy(data[1:], []byte(db))
	if err := c.writePacket(data); err != nil {
		return err
	}
	if err := c.flush(); err != nil {
		return err
	}
	return nil
}

func (c *Conn) readColumnDefinition(field *querypb.Field, index int) error {
	colDef, err := c.ReadPacket()
	if err != nil {
		return err
	}

	// Catalog is ignored, always set to "def"
	pos, ok := skipLenEncString(colDef, 0)
	if !ok {
		return fmt.Errorf("skipping col %v catalog failed", index)
	}

	// schema, table, orgTable, name and OrgName are strings.
	field.Database, pos, ok = readLenEncString(colDef, pos)
	if !ok {
		return fmt.Errorf("extracting col %v schema failed", index)
	}
	field.Table, pos, ok = readLenEncString(colDef, pos)
	if !ok {
		return fmt.Errorf("extracting col %v table failed", index)
	}
	field.OrgTable, pos, ok = readLenEncString(colDef, pos)
	if !ok {
		return fmt.Errorf("extracting col %v org_table failed", index)
	}
	field.Name, pos, ok = readLenEncString(colDef, pos)
	if !ok {
		return fmt.Errorf("extracting col %v name failed", index)
	}
	field.OrgName, pos, ok = readLenEncString(colDef, pos)
	if !ok {
		return fmt.Errorf("extracting col %v org_name failed", index)
	}

	// Skip length of fixed-length fields.
	pos++

	// characterSet is a uint16.
	characterSet, pos, ok := readUint16(colDef, pos)
	if !ok {
		return fmt.Errorf("extracting col %v characterSet failed", index)
	}
	field.Charset = uint32(characterSet)

	// columnLength is a uint32.
	field.ColumnLength, pos, ok = readUint32(colDef, pos)
	if !ok {
		return fmt.Errorf("extracting col %v columnLength failed", index)
	}

	// type is one byte
	t, pos, ok := readByte(colDef, pos)
	if !ok {
		return fmt.Errorf("extracting col %v type failed", index)
	}

	// flags is 2 bytes
	flags, pos, ok := readUint16(colDef, pos)
	if !ok {
		return fmt.Errorf("extracting col %v flags failed", index)
	}
	field.Flags = uint32(flags)

	// Convert MySQL type to Vitess type.
	field.Type, err = sqltypes.MySQLToType(int64(t), int64(flags))
	if err != nil {
		return fmt.Errorf("MySQLToType(%v,%v) failed for column %v: %v", t, flags, index, err)
	}

	// Decimals is a byte.
	decimals, pos, ok := readByte(colDef, pos)
	if !ok {
		return fmt.Errorf("extracting col %v decimals failed", index)
	}
	field.Decimals = uint32(decimals)

	return nil
}

// readColumnDefinitionType is a faster version of
// readColumnDefinition that only fills in the Type.
func (c *Conn) readColumnDefinitionType(field *querypb.Field, index int) error {
	colDef, err := c.ReadPacket()
	if err != nil {
		return err
	}

	// catalog, schema, table, orgTable, name and orgName are
	// strings, all skipped.
	pos, ok := skipLenEncString(colDef, 0)
	if !ok {
		return fmt.Errorf("skipping col %v catalog failed", index)
	}
	pos, ok = skipLenEncString(colDef, pos)
	if !ok {
		return fmt.Errorf("skipping col %v schema failed", index)
	}
	pos, ok = skipLenEncString(colDef, pos)
	if !ok {
		return fmt.Errorf("skipping col %v table failed", index)
	}
	pos, ok = skipLenEncString(colDef, pos)
	if !ok {
		return fmt.Errorf("skipping col %v org_table failed", index)
	}
	pos, ok = skipLenEncString(colDef, pos)
	if !ok {
		return fmt.Errorf("skipping col %v name failed", index)
	}
	pos, ok = skipLenEncString(colDef, pos)
	if !ok {
		return fmt.Errorf("skipping col %v org_name failed", index)
	}

	// Skip length of fixed-length fields.
	pos++

	// characterSet is a uint16.
	_, pos, ok = readUint16(colDef, pos)
	if !ok {
		return fmt.Errorf("extracting col %v characterSet failed", index)
	}

	// columnLength is a uint32.
	_, pos, ok = readUint32(colDef, pos)
	if !ok {
		return fmt.Errorf("extracting col %v columnLength failed", index)
	}

	// type is one byte
	t, pos, ok := readByte(colDef, pos)
	if !ok {
		return fmt.Errorf("extracting col %v type failed", index)
	}

	// flags is 2 bytes
	flags, pos, ok := readUint16(colDef, pos)
	if !ok {
		return fmt.Errorf("extracting col %v flags failed", index)
	}

	// Convert MySQL type to Vitess type.
	field.Type, err = sqltypes.MySQLToType(int64(t), int64(flags))
	if err != nil {
		return fmt.Errorf("MySQLToType(%v,%v) failed for column %v: %v", t, flags, index, err)
	}

	// skip decimals

	return nil
}

func (c *Conn) parseRow(data []byte, fields []*querypb.Field) ([]sqltypes.Value, error) {
	colNumber := len(fields)
	result := make([]sqltypes.Value, colNumber)
	pos := 0
	for i := 0; i < colNumber; i++ {
		if data[pos] == 0xfb {
			pos++
			continue
		}
		var s []byte
		var ok bool
		s, pos, ok = readLenEncStringAsBytes(data, pos)
		if !ok {
			return nil, fmt.Errorf("decoding string failed")
		}
		result[i] = sqltypes.MakeTrusted(fields[i].Type, s)
	}
	return result, nil
}

// ExecuteFetch is the same as sqldb.Conn.ExecuteFetch.
func (c *Conn) ExecuteFetch(query string, maxrows int, wantfields bool) (*sqltypes.Result, error) {
	// This is a new command, need to reset the sequence.
	c.sequence = 0

	// Send the query as a COM_QUERY packet.
	if err := c.writeComQuery(query); err != nil {
		return nil, err
	}

	// Get the result.
	affectedRows, lastInsertID, colNumber, err := c.readComQueryResponse()
	if err != nil {
		return nil, err
	}
	if colNumber == 0 {
		// OK packet, means no results. Just use the numbers.
		return &sqltypes.Result{
			RowsAffected: affectedRows,
			InsertID:     lastInsertID,
		}, nil
	}

	fields := make([]querypb.Field, colNumber)
	result := &sqltypes.Result{
		Fields: make([]*querypb.Field, colNumber),
	}

	// Read column headers. One packet per column.
	// Build the fields.
	for i := 0; i < colNumber; i++ {
		result.Fields[i] = &fields[i]

		if wantfields {
			if err := c.readColumnDefinition(result.Fields[i], i); err != nil {
				return nil, err
			}
		} else {
			if err := c.readColumnDefinitionType(result.Fields[i], i); err != nil {
				return nil, err
			}
		}
	}

	if c.Capabilities&CapabilityClientDeprecateEOF == 0 {
		// EOF is only present here if it's not deprecated.
		data, err := c.ReadPacket()
		if err != nil {
			return nil, err
		}
		switch data[0] {
		case EOFPacket:
			// This is what we expect.
			// Warnings and status flags are ignored.
			break
		case ErrPacket:
			// Error packet.
			return nil, parseErrorPacket(data)
		default:
			return nil, fmt.Errorf("unexpected packet after fields: %v", data)
		}
	}

	// read each row until EOF or OK packet.
	for {
		data, err := c.ReadPacket()
		if err != nil {
			return nil, err
		}

		switch data[0] {
		case EOFPacket:
			// This packet may be one of two kinds:
			// - an EOF packet,
			// - an OK packet with an EOF header if
			// CapabilityClientDeprecateEOF is set.
			// We do not parse it anyway, so it doesn't matter.

			// Strip the partial Fields before returning.
			if !wantfields {
				result.Fields = nil
			}
			return result, nil
		case ErrPacket:
			// Error packet.
			return nil, parseErrorPacket(data)
		}

		// Check we're not over the limit before we add more.
		if len(result.Rows) == maxrows {
			if err := c.drainResults(); err != nil {
				return nil, err
			}
			return nil, &sqldb.SQLError{
				Num:     0,
				Message: fmt.Sprintf("Row count exceeded %d", maxrows),
				Query:   query,
			}
		}

		// Regular row.
		row, err := c.parseRow(data, result.Fields)
		if err != nil {
			return nil, err
		}
		result.Rows = append(result.Rows, row)
	}
}

// drainResults will read all packets for a result set and ignore them.
func (c *Conn) drainResults() error {
	for {
		data, err := c.ReadPacket()
		if err != nil {
			return err
		}

		switch data[0] {
		case EOFPacket:
			// This packet may be one of two kinds:
			// - an EOF packet,
			// - an OK packet with an EOF header if
			// CapabilityClientDeprecateEOF is set.
			// We do not parse it anyway, so it doesn't matter.
			return nil
		case ErrPacket:
			// Error packet.
			return parseErrorPacket(data)
		}
	}
}

func (c *Conn) readComQueryResponse() (uint64, uint64, int, error) {
	data, err := c.ReadPacket()
	if err != nil {
		return 0, 0, 0, err
	}
	if len(data) == 0 {
		return 0, 0, 0, fmt.Errorf("invalid empty COM_QUERY response packet")
	}

	switch data[0] {
	case OKPacket:
		affectedRows, lastInsertID, _, _, err := parseOKPacket(data)
		return affectedRows, lastInsertID, 0, err
	case ErrPacket:
		// Error
		return 0, 0, 0, parseErrorPacket(data)
	case 0xfb:
		// Local infile
		return 0, 0, 0, fmt.Errorf("not implemented")
	}

	n, pos, ok := readLenEncInt(data, 0)
	if !ok {
		return 0, 0, 0, fmt.Errorf("cannot get column number")
	}
	if pos != len(data) {
		return 0, 0, 0, fmt.Errorf("extra data in COM_QUERY response")
	}
	return 0, 0, int(n), nil
}

func (c *Conn) parseComQuery(data []byte) string {
	return string(data[1:])
}

func (c *Conn) parseComInitDB(data []byte) string {
	return string(data[1:])
}

func (c *Conn) sendColumnCount(count uint64) error {
	length := lenEncIntSize(count)
	data := make([]byte, length)
	writeLenEncInt(data, 0, count)
	return c.writePacket(data)
}

func (c *Conn) writeColumnDefinition(field *querypb.Field) error {
	length := 4 + // lenEncStringSize("def")
		lenEncStringSize(field.Database) +
		lenEncStringSize(field.Table) +
		lenEncStringSize(field.OrgTable) +
		lenEncStringSize(field.Name) +
		lenEncStringSize(field.OrgName) +
		1 + // length of fixed length fields
		2 + // character set
		4 + // column length
		1 + // type
		2 + // flags
		1 + // decimals
		2 // filler

	// Only get the type back. The flags can be retrieved from the
	// Field.
	typ, _ := sqltypes.TypeToMySQL(field.Type)

	data := make([]byte, length)
	pos := 0

	pos = writeLenEncString(data, pos, "def") // Always the same.
	pos = writeLenEncString(data, pos, field.Database)
	pos = writeLenEncString(data, pos, field.Table)
	pos = writeLenEncString(data, pos, field.OrgTable)
	pos = writeLenEncString(data, pos, field.Name)
	pos = writeLenEncString(data, pos, field.OrgName)
	pos = writeByte(data, pos, 0x0c)
	pos = writeUint16(data, pos, uint16(field.Charset))
	pos = writeUint32(data, pos, field.ColumnLength)
	pos = writeByte(data, pos, byte(typ))
	pos = writeUint16(data, pos, uint16(field.Flags))
	pos = writeByte(data, pos, byte(field.Decimals))
	pos += 2

	if pos != len(data) {
		return fmt.Errorf("internal error: packing of column definition used %v bytes instead of %v", pos, len(data))
	}

	return c.writePacket(data)
}

func (c *Conn) writeRow(row []sqltypes.Value) error {
	length := 0
	for _, val := range row {
		if val.IsNull() {
			length++
		} else {
			l := len(val.Raw())
			length += lenEncIntSize(uint64(l)) + l
		}
	}

	data := make([]byte, length)
	pos := 0
	for _, val := range row {
		if val.IsNull() {
			pos = writeByte(data, pos, NullValue)
		} else {
			l := len(val.Raw())
			pos = writeLenEncInt(data, pos, uint64(l))
			pos += copy(data[pos:], val.Raw())
		}
	}

	if pos != length {
		return fmt.Errorf("internal error packet row: got %v bytes but expected %v", pos, length)
	}

	return c.writePacket(data)
}

// writeResult writes a query Result to the wire.
func (c *Conn) writeResult(result *sqltypes.Result) error {
	if len(result.Fields) == 0 {
		// This is just an INSERT result, send an OK packet.
		return c.writeOKPacket(result.RowsAffected, result.InsertID, c.StatusFlags, 0)
	}

	// Now send a packet with just the number of fields.
	if err := c.sendColumnCount(uint64(len(result.Fields))); err != nil {
		return err
	}

	// Now send each Field.
	for _, field := range result.Fields {
		if err := c.writeColumnDefinition(field); err != nil {
			return err
		}
	}

	// Now send an EOF packet.
	if c.Capabilities&CapabilityClientDeprecateEOF == 0 {
		// With CapabilityClientDeprecateEOF, we do not send this EOF.
		if err := c.writeEOFPacket(c.StatusFlags, 0); err != nil {
			return err
		}
	}

	// Now send one packet per row.
	for _, row := range result.Rows {
		if err := c.writeRow(row); err != nil {
			return err
		}
	}

	// And send either an EOF, or an OK packet.
	// FIXME(alainjobart) if multi result is set, can send more after this.
	// See doc.go.
	if c.Capabilities&CapabilityClientDeprecateEOF == 0 {
		if err := c.writeEOFPacket(c.StatusFlags, 0); err != nil {
			return err
		}
		if err := c.flush(); err != nil {
			return err
		}
	} else {
		// This will flush too.
		if err := c.writeOKPacketWithEOFHeader(0, 0, c.StatusFlags, 0); err != nil {
			return err
		}
	}

	return nil
}
