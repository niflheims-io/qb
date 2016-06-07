package qb

func (self *Tx) exec(ql string, lines [][]field) (int64, error) {
	lineLen := len(lines)
	if lineLen == 0 {
		return 0, nil
	}
	stmt, stmtErr := self.tx.Prepare(ql)
	if stmtErr != nil {
		return 0, stmtErr
	}
	defer stmt.Close()
	totalRsAffected := int64(0)
	for i := 0 ; i < lineLen ; i ++ {
		line := lines[i]
		fieldLen := len(line)
		args := make([]interface{},0,len(line))
		for j := 0 ; j < fieldLen ; j ++  {
			field := line[j]
			args = append(args, field.value)
		}
		rs, execErr := stmt.Exec(args...)
		if execErr != nil {
			return 0, execErr
		}
		rsAffected, rsErr := rs.RowsAffected()
		if rsErr != nil {
			return 0, rsErr
		}
		totalRsAffected = totalRsAffected + rsAffected
	}
	return totalRsAffected, nil
}
