package engines

import "github.com/towgo/towgo/dao/datatable/models"

type Gorm struct{}

func (g *Gorm) CreateTable(table *models.TableModel) (bool, error) {
	return true, nil
}
func (g *Gorm) AddField(table *models.TableFieldModel) (bool, error) {
	return true, nil
}

func (g *Gorm) DeleteField(table *models.TableFieldModel) (bool, error) {
	return true, nil
}
func (g *Gorm) UpdateTable(table *models.TableModel) (bool, error) {
	return true, nil
}
func (g *Gorm) DropTable(table *models.TableModel) (bool, error) {
	return true, nil
}
func (g *Gorm) Create(tableModel *models.TableModel) (int64, error) {
	return 0, nil
}
func (g *Gorm) Update(tableModel *models.TableModel) error {
	return nil
}
func (g *Gorm) Delete(tableModel *models.TableModel) error {
	return nil
}
func (g *Gorm) Query(tableQuery *models.TableQuery) (interface{}, error) {
	return map[string]interface{}{}, nil
}
