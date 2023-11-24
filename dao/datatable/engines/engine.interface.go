package engines

import "github.com/towgo/towgo/dao/datatable/models"

type TableOrmEngine interface {
	CreateTable(table *models.TableModel) (bool, error)

	AddField(table *models.TableFieldModel) (bool, error)

	DeleteField(table *models.TableFieldModel) (bool, error)

	UpdateTable(table *models.TableModel) (bool, error)

	DropTable(table *models.TableModel) (bool, error)

	Create(tableModel *models.TableModel) (int64, error)

	Update(tableModel *models.TableModel) error

	Delete(tableModel *models.TableModel) error

	Query(tableQuery *models.TableQuery) (interface{}, error)
}
