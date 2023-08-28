package gormq

import (
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	"github.com/glothriel/grf/pkg/queries/crud"
	"gorm.io/gorm"
)

type CreateTxHook func(ctx *gin.Context, iv models.InternalValue, db *gorm.DB) (models.InternalValue, error)
type UpdateTxHook func(ctx *gin.Context, old models.InternalValue, new models.InternalValue, id any, db *gorm.DB) (models.InternalValue, error)
type DestroyTxHook func(ctx *gin.Context, id any, db *gorm.DB) error

type CreateTxHooks struct {
	before CreateTxHook
	after  CreateTxHook
}

type UpdateTxHooks struct {
	before UpdateTxHook
	after  UpdateTxHook
}

type DestroyTxHooks struct {
	before DestroyTxHook
	after  DestroyTxHook
}

func AfterCreate(hook CreateTxHook) CreateTxHooks {
	return CreateTxHooks{after: hook}
}

func BeforeCreate(hook CreateTxHook) CreateTxHooks {
	return CreateTxHooks{before: hook}
}

func AfterUpdate(hook UpdateTxHook) UpdateTxHooks {
	return UpdateTxHooks{after: hook}
}

func BeforeUpdate(hook UpdateTxHook) UpdateTxHooks {
	return UpdateTxHooks{before: hook}
}

func AfterDestroy(hook DestroyTxHook) DestroyTxHooks {
	return DestroyTxHooks{after: hook}
}

func BeforeDestroy(hook DestroyTxHook) DestroyTxHooks {
	return DestroyTxHooks{before: hook}
}

func CreateTx(hooks ...CreateTxHooks) func(crud.CreateQueryFunc) crud.CreateQueryFunc {
	return func(previous crud.CreateQueryFunc) crud.CreateQueryFunc {
		return func(ctx *gin.Context, new models.InternalValue) (models.InternalValue, error) {
			var childResult models.InternalValue
			previousQuery := CtxQuery(ctx)
			defer CtxInitQuery(ctx)
			if txErr := previousQuery.Transaction(func(tx *gorm.DB) error {
				CtxSetQuery(ctx, tx)
				var createdIV = new
				var childErr error

				for _, hook := range hooks {
					if hook.before != nil {
						if createdIV, childErr = hook.before(ctx, createdIV, tx); childErr != nil {
							return childErr
						}
					}
				}

				childResult, childErr = previous(ctx, createdIV)
				if childErr != nil {
					return childErr
				}

				for _, hook := range hooks {
					if hook.after != nil {
						if childResult, childErr = hook.after(ctx, childResult, tx); childErr != nil {
							return childErr
						}
					}
				}

				return nil
			}); txErr != nil {
				return nil, txErr
			}
			return childResult, nil
		}
	}
}

func UpdateTx(hooks ...UpdateTxHooks) func(crud.UpdateQueryFunc) crud.UpdateQueryFunc {
	return func(previous crud.UpdateQueryFunc) crud.UpdateQueryFunc {
		return func(ctx *gin.Context, old models.InternalValue, new models.InternalValue, id any) (models.InternalValue, error) {
			var childResult models.InternalValue
			previousQuery := CtxQuery(ctx)
			defer CtxInitQuery(ctx)
			if txErr := previousQuery.Transaction(func(tx *gorm.DB) error {
				CtxSetQuery(ctx, tx)
				var updatedIV = new
				var childErr error

				for _, hook := range hooks {
					if hook.before != nil {
						if updatedIV, childErr = hook.before(ctx, old, updatedIV, id, tx); childErr != nil {
							return childErr
						}
					}
				}

				childResult, childErr = previous(ctx, old, updatedIV, id)
				if childErr != nil {
					return childErr
				}

				for _, hook := range hooks {
					if hook.after != nil {
						if childResult, childErr = hook.after(ctx, old, childResult, id, tx); childErr != nil {
							return childErr
						}
					}
				}

				return nil
			}); txErr != nil {
				return nil, txErr
			}

			return childResult, nil
		}
	}
}

func DestroyTx(hooks ...DestroyTxHooks) func(crud.DestroyQueryFunc) crud.DestroyQueryFunc {
	return func(previous crud.DestroyQueryFunc) crud.DestroyQueryFunc {
		return func(ctx *gin.Context, id any) error {
			previousQuery := CtxQuery(ctx)
			defer CtxInitQuery(ctx)
			return previousQuery.Transaction(func(tx *gorm.DB) error {
				CtxSetQuery(ctx, tx)
				var childErr error

				for _, hook := range hooks {
					if hook.before != nil {
						if childErr = hook.before(ctx, id, tx); childErr != nil {
							return childErr
						}
					}
				}

				childErr = previous(ctx, id)
				if childErr != nil {
					return childErr
				}

				for _, hook := range hooks {
					if hook.after != nil {
						if childErr = hook.after(ctx, id, tx); childErr != nil {
							return childErr
						}
					}
				}

				return nil
			})
		}
	}
}
