package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/FloatTech/zbputils/control/web/types"
	"github.com/FloatTech/zbputils/job"
)

// JobList 任务列表
//
//	@Tags			任务
//	@Summary		任务列表
//	@Description	任务列表
//	@Router			/api/job/list [get]
//	@Success		200	{object}	types.Response{result=[]job.Job}	"成功"
func JobList(context *gin.Context) {
	rsp, err := job.List()
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       rsp,
		Message:      "",
		ResponseType: "ok",
	})
}

// JobAdd 添加任务
//
//	@Tags			任务
//	@Summary		添加任务
//	@Description	添加任务
//	@Router			/api/job/add [post]
//	@Param			object	body		job.Job				false	"添加任务入参"
//	@Success		200		{object}	types.Response	"成功"
func JobAdd(context *gin.Context) {
	var (
		j job.Job
	)
	err := context.ShouldBind(&j)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	err = job.Add(&j)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       nil,
		Message:      "",
		ResponseType: "ok",
	})
}

// JobDelete 删除任务
//
//	@Tags			任务
//	@Summary		删除任务
//	@Description	删除任务
//	@Router			/api/job/delete [post]
//	@Param			object	body		job.DeleteReq		false	"删除任务的入参"
//	@Success		200		{object}	types.Response	"成功"
func JobDelete(context *gin.Context) {
	var (
		req job.DeleteReq
	)
	err := context.ShouldBind(&req)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	err = job.Delete(&req)
	if err != nil {
		context.JSON(http.StatusOK, types.Response{
			Code:         -1,
			Result:       nil,
			Message:      err.Error(),
			ResponseType: "error",
		})
		return
	}
	context.JSON(http.StatusOK, types.Response{
		Code:         0,
		Result:       nil,
		Message:      "",
		ResponseType: "ok",
	})
}
