package db

import (
	"context"
	"errors"
	"pipbot/graph"
	"pipbot/graph/model"
)

var _ graph.MutationResolver = (*Client)(nil)
var _ graph.QueryResolver = (*Client)(nil)

type Client struct {
	*PrismaClient
}

func ConvertMatrix(m *MatrixModel) (*model.Matrix, error) {
	grids := m.Grids()
	ret := &model.Matrix{
		ID:    m.ID,
		Name:  m.Name,
		Grids: make([]*model.Grid, len(grids)),
	}
	for i, g := range grids {
		grid, err := ConvertGrid(&g)
		if err != nil {
			return nil, err
		}
		ret.Grids[i] = grid
	}
	return ret, nil
}

func (c *Client) Matrices(ctx context.Context) ([]*model.Matrix, error) {
	matrices, err := c.Matrix.FindMany().Exec(ctx)
	if err != nil {
		return nil, err
	}
	if len(matrices) == 0 {
		return []*model.Matrix{}, nil
	}
	result := make([]*model.Matrix, len(matrices))
	for i, m := range matrices {
		mat, err := ConvertMatrix(&m)
		if err != nil {
			return nil, err
		}
		result[i] = mat
	}
	return result, nil
}

func ConvertGrid(g *GridModel) (*model.Grid, error) {
	gridPos, found := g.Home()
	if !found {
		return nil, errors.New("grid has no home position")
	}
	return &model.Grid{
		ID:   g.ID,
		Name: g.Name,
		Home: &model.Position{
			X: gridPos.X,
			Y: gridPos.Y,
			Z: gridPos.Z,
		},
		RowSpace: g.RowSpace,
		ColSpace: g.ColSpace,
		NRows:    g.NRows,
		NCols:    g.NCols,
	}, nil
}

func (c *Client) Grids(ctx context.Context, matrixID string) ([]*model.Grid, error) {
	grids, err := c.Grid.FindMany(Grid.MatrixID.Equals(matrixID)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	if len(grids) == 0 {
		return []*model.Grid{}, nil
	}
	result := make([]*model.Grid, len(grids))
	for i, g := range grids {
		v, err := ConvertGrid(&g)
		if err != nil {
			return nil, err
		}
		result[i] = v
	}
	return result, nil
}

func ConvertTransfer(t *TransferModel) (*model.Transfer, error) {
	ret := &model.Transfer{
		ID: t.ID,
		Source: &model.Node{
			Grid:     t.SourceGrid,
			Position: t.SourcePosition,
			Aspirate: t.SourceAspirate,
		},
		Dest: &model.Node{
			Grid:     t.DestGrid,
			Position: t.DestPosition,
			Aspirate: t.DestAspirate,
		},
		Volume: t.Volume,
	}
	if name, exists := t.Name(); exists {
		ret.Name = &name
	}
	if group, exists := t.Group(); exists {
		ret.Group = &group
	}
	return ret, nil
}

func ConvertRecipe(r *RecipeModel) (*model.Recipe, error) {
	mat, err := ConvertMatrix(r.Matrix())
	if err != nil {
		return nil, err
	}
	transfers := r.Transfers()
	ret := &model.Recipe{
		ID:        r.ID,
		Name:      r.Name,
		Matrix:    mat,
		Transfers: make([]*model.Transfer, len(transfers)),
	}
	for i, t := range transfers {
		transfer, err := ConvertTransfer(&t)
		if err != nil {
			return nil, err
		}
		ret.Transfers[i] = transfer
	}
	if url, exists := r.DownloadURL(); exists {
		ret.Download = url
	}

	return ret, nil
}

func (c *Client) Recipes(ctx context.Context) ([]*model.Recipe, error) {
	recipes, err := c.PrismaClient.Recipe.FindMany().Exec(ctx)
	if err != nil {
		return nil, err
	}
	if len(recipes) == 0 {
		return []*model.Recipe{}, nil
	}
	result := make([]*model.Recipe, len(recipes))
	for i, r := range recipes {
		recipe, err := ConvertRecipe(&r)
		if err != nil {
			return nil, err
		}
		result[i] = recipe
	}
	return result, nil
}

func (c *Client) Recipe(ctx context.Context, id string) (*model.Recipe, error) {
	recipe, err := c.PrismaClient.Recipe.FindUnique(Recipe.ID.Equals(id)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return ConvertRecipe(recipe)
}

func (c *Client) CreateMatrix(ctx context.Context, matrix model.NewMatrix) (*model.Matrix, error) {
	mat, err := c.Matrix.CreateOne(
		Matrix.Name.Set(matrix.Name),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}

	ret, err := ConvertMatrix(mat)
	if err != nil {
		return nil, err
	}
	if matrix.Grids == nil {
		return ret, nil
	}
	if len(matrix.Grids) == 0 {
		return ret, nil
	}
	ret.Grids = make([]*model.Grid, len(matrix.Grids))
	for i, grid := range matrix.Grids {
		g, err := c.AddGrid(ctx, mat.ID, *grid)
		if err != nil {
			return nil, err
		}
		ret.Grids[i] = g
	}
	return ret, err
}

func ConvertPosition(p *PositionModel) *model.Position {
	return &model.Position{
		X: p.X,
		Y: p.Y,
		Z: p.Z,
	}
}

func (c *Client) AddGrid(ctx context.Context, matrixID string, grid model.NewGrid) (*model.Grid, error) {
	g, err := c.Grid.CreateOne(
		Grid.Name.Set(grid.Name),
		Grid.Matrix.Link(
			Matrix.ID.Equals(matrixID),
		),
		Grid.RowSpace.Set(grid.RowSpace),
		Grid.ColSpace.Set(grid.ColSpace),
		Grid.NRows.Set(grid.NRows),
		Grid.NCols.Set(grid.NCols),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}

	ret, err := ConvertGrid(g)
	if err != nil {
		return nil, err
	}

	pos, err := c.Position.CreateOne(
		Position.X.Set(grid.Home.X),
		Position.Y.Set(grid.Home.Y),
		Position.Z.Set(grid.Home.Z),
		Position.Grid.Link(
			Grid.ID.Equals(g.ID),
		),
	).Exec(ctx)

	if err != nil {
		return nil, err
	}

	ret.Home = ConvertPosition(pos)

	return ret, nil
}

func (c *Client) CreateRecipe(ctx context.Context, recipe model.NewRecipe) (*model.Recipe, error) {
	r, err := c.PrismaClient.Recipe.CreateOne(
		Recipe.Name.Set(recipe.Name),
		Recipe.Matrix.Link(
			Matrix.ID.Equals(recipe.MatrixID),
		),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	ret, err := ConvertRecipe(r)
	if err != nil {
		return nil, err
	}
	if recipe.Transfers == nil {
		return ret, nil
	}
	if len(recipe.Transfers) == 0 {
		return ret, nil
	}
	ret.Transfers = make([]*model.Transfer, len(recipe.Transfers))
	for i, transfer := range recipe.Transfers {
		t, err := c.AddTransfer(ctx, r.ID, *transfer)
		if err != nil {
			return nil, err
		}
		ret.Transfers[i] = t
	}
	return ret, nil
}

func (c *Client) AddTransfer(ctx context.Context, recipeID string, transfer model.NewTransfer) (*model.Transfer, error) {
	aspirateSource := false
	if transfer.Source.Aspirate != nil {
		aspirateSource = *transfer.Source.Aspirate
	}
	aspirateDest := false
	if transfer.Dest.Aspirate != nil {
		aspirateDest = *transfer.Dest.Aspirate
	}
	t, err := c.Transfer.CreateOne(
		Transfer.SampleID.Set(transfer.SampleID),
		Transfer.SourceGrid.Set(transfer.Source.Grid),
		Transfer.SourcePosition.Set(transfer.Source.Position),
		Transfer.SourceAspirate.Set(aspirateSource),
		Transfer.DestGrid.Set(transfer.Dest.Grid),
		Transfer.DestPosition.Set(transfer.Dest.Position),
		Transfer.DestAspirate.Set(aspirateDest),
		Transfer.Volume.Set(transfer.Volume),
		Transfer.Recipe.Link(
			Recipe.ID.Equals(recipeID),
		),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return ConvertTransfer(t)
}
