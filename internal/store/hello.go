package store

import "context"

type HellResult struct {
	Data string `db:"data"`
}

func (d *daoImpl) Hello(ctx context.Context, name string) (string, error) {
	//ret := &HellResult{}
	//err := d.sqlRepo.QuerySingle(ctx, &ret, `SELECT CONCAT("Hello ",?) as "data";`, name)
	//if err != nil {
	//	return "", err
	//}

	return "Hello " + name, nil
}
