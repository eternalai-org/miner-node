package watcher

import (
	"bytes"
	"encoding/json"
	"errors"
	"eternal-infer-worker/libs/eaimodel"
	"eternal-infer-worker/libs/lighthouse"
	"eternal-infer-worker/types"
	"fmt"
	"log"
)

func (tskw *TaskWatcher) executeWorkerTask(task *types.TaskInfo) error {
	newRunner := tskw.GetRunner(task.TaskID)
	if newRunner == nil {
		log.Println("runner not found", task.TaskID)
		return errors.New("runner not found")
	}

	modelInst, err := tskw.modelManager.GetModelInstance(task.ModelContract)
	if err != nil {
		log.Println("get model instance error: ", err)
		return err
	}

	finalResult := &bytes.Buffer{}
	ext := "png"
	switch modelInst.ModelInfo.Metadata.ModelType {
	case eaimodel.ModelTypeImage:
		ext = "png"
	case eaimodel.ModelTypeText:
		ext = "txt"
	}
	output := fmt.Sprintf("/output/%v.%v", task.TaskID, ext)

	err = newRunner.Run(output)
	if err != nil {
		log.Println("run task error: ", err)
		return err
	}

	result := fmt.Sprintf("%v/%v.%v", modelInst.ResultDir, task.TaskID, ext)
	resultData, err := readResultFile(result)
	if err != nil {
		log.Println("read result file error: ", err)
		return err
	}

	switch modelInst.ModelInfo.Metadata.ModelType {
	case eaimodel.ModelTypeImage:
		finalResult = bytes.NewBuffer(resultData)
	case eaimodel.ModelTypeText:
		finalResult = bytes.NewBuffer(resultData)
		ext = "txt"
	}

	log.Println("uploading result: ", fmt.Sprintf("%v_result.%v", task.TaskID, ext))

	cid, err := lighthouse.UploadData(tskw.lighthouseAPI, fmt.Sprintf("%v_result.%v", task.TaskID, ext), finalResult.Bytes())
	if err != nil {
		log.Println("upload data error: ", err)
		return err
	}
	resultLink := fmt.Sprintf("ipfs://%v", cid)

	taskResult := eaimodel.TaskResult{
		ResultURI: resultLink,
	}

	resultData, err = json.Marshal(taskResult)
	if err != nil {
		log.Println("marshal result error: ", err)
		return err
	}

	log.Printf("\nsubmitting result for task %v size %v\n", task.TaskID, len(resultData))

	err = tskw.SubmitResult(task.AssignmentID, resultData)
	if err != nil {
		log.Println("submit result error: ", err)
		return err
	}

	return nil
}
