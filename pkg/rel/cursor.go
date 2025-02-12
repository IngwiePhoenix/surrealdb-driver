package rel

/*
// TODO: Do I need this?
var _ (rel.Cursor) = (*Cursor)(nil)

type Cursor struct {
	*driver.SurrealRows
}

func (c *Cursor) Close() error {
	return c.SurrealRows.Close()
}

func (c *Cursor) Fields() ([]string, error) {
	return c.SurrealRows.Columns(), nil
}

func (c *Cursor) Next() bool {
	// TODO: Proper error handling
	res := c.SurrealRows.Next()
	return res == nil
}

func (c *Cursor) Scan(dest ...any) bool {

}
func (c *Cursor) NopScanner() any {
	// TODO: NOP
	return nil
}
*/
