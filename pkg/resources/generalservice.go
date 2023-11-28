// Copyright 2020 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"yunion.io/x/onecloud/pkg/mcclient/modules"

	"yunion.io/x/jsonutils"
	onecloudv1 "yunion.io/x/onecloud-service-operator/api/v1"
)

var (
	RequestGeneralService = Request.Resource(ResourceGeneralService)
)

func init() {
	Register(ResourceGeneralService, modules.GeneralServiceProviderManager)
}

type GeneralService struct {
	GeneralService *onecloudv1.GeneralService
	logger         logr.Logger
}

func NewGeneralService(gs *onecloudv1.GeneralService, logger logr.Logger) GeneralService {
	return GeneralService{gs, logger}
}

func (gs GeneralService) GetIResource() onecloudv1.IResource {
	return gs.GeneralService
}

func (gs GeneralService) GetResourceName() Resource {
	return ResourceGeneralService
}

func (gs GeneralService) Create(ctx context.Context, data interface{}) (onecloudv1.ExternalInfoBase, error) {
	params, ok := data.(map[string]interface{})
	if !ok {
		return onecloudv1.ExternalInfoBase{}, errors.New("invalid params")
	}
	params["releaseId"] = gs.GeneralService.Spec.ReleaseId

	service, _, err := RequestGeneralService.Operation(OperGet).Apply(ctx, gs.GeneralService.Spec.ServiceId, nil)
	if err != nil {
		return onecloudv1.ExternalInfoBase{}, err
	}

	ret, extInfo, err := RequestGeneralService.Operation(GeneralServiceCreate).Apply(ctx, gs.GeneralService.Spec.ServiceId, jsonutils.Marshal(data).(*jsonutils.JSONDict))
	if err != nil {
		return onecloudv1.ExternalInfoBase{}, err
	}

	extInfo.Id, _ = ret.GetString("id")
	extInfo.Status, _ = ret.GetString("release_status")
	gs.GeneralService.Status.ExternalInfo.ExternalId, _ = ret.GetString("external_id")
	gs.GeneralService.Status.ExternalInfo.PrimaryKey, _ = service.GetString("primary_key")

	return extInfo, nil
}

func (gs GeneralService) Update(ctx context.Context, params map[string]interface{}) (*onecloudv1.GeneralServiceStatus, error) {
	params["id"] = gs.GeneralService.Status.ExternalInfo.Id

	ret, extInfo, err := RequestGeneralService.Operation(GeneralServiceUpdate).Apply(ctx, gs.GeneralService.Spec.ServiceId, jsonutils.Marshal(params).(*jsonutils.JSONDict))
	if err != nil {
		return nil, err
	}
	extInfo.Id, _ = ret.GetString("id")
	extInfo.Status, _ = ret.GetString("release_status")

	status := gs.GeneralService.Status.DeepCopy()
	status.Phase = onecloudv1.ResourcePending
	status.Reason = fmt.Sprintf("Exec '%s' successfully", extInfo.Action)
	status.ExternalInfo.ExternalInfoBase = extInfo

	return status, nil
}

func (gs GeneralService) Delete(ctx context.Context) (onecloudv1.ExternalInfoBase, error) {
	params := jsonutils.NewDict()
	params.Set("id", jsonutils.NewString(gs.GeneralService.Status.ExternalInfo.Id))

	ret, extInfo, err := RequestGeneralService.Operation(GeneralServiceDelete).Apply(ctx, gs.GeneralService.Spec.ServiceId, params)
	if err != nil {
		return onecloudv1.ExternalInfoBase{}, err
	}

	extInfo.Id, _ = ret.GetString("id")
	extInfo.Status, _ = ret.GetString("release_status")

	return extInfo, err
}

func (gs GeneralService) GetDetail(ctx context.Context) (map[string]interface{}, error) {
	params := jsonutils.NewDict()
	params.Set("id", jsonutils.NewString(gs.GeneralService.Status.ExternalInfo.Id))

	ret, _, err := RequestGeneralService.Operation(GeneralServiceGet).Apply(ctx, gs.GeneralService.Spec.ServiceId, params)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(ret.String()), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (gs GeneralService) GetStatus(ctx context.Context) (onecloudv1.IResourceStatus, error) {
	params := jsonutils.NewDict()
	params.Set("id", jsonutils.NewString(gs.GeneralService.Status.ExternalInfo.Id))

	service, _, err := RequestGeneralService.Operation(OperGet).Apply(ctx, gs.GeneralService.Spec.ServiceId, nil)
	if err != nil {
		return nil, err
	}

	status := gs.GeneralService.Status.DeepCopy()
	ret, extInfo, err := RequestGeneralService.Operation(GeneralServiceGet).Apply(ctx, gs.GeneralService.Spec.ServiceId, params)
	if err != nil {
		return nil, err
	}

	extInfo.Id, _ = ret.GetString("id")
	extInfo.Status, _ = ret.GetString("release_status")
	status.ExternalInfo.ExternalId, _ = ret.GetString("external_id")
	status.ExternalInfo.PrimaryKey, _ = service.GetString("primary_key")
	status.ExternalInfo.ExternalInfoBase = extInfo

	switch extInfo.Status {
	case "deploy_failed":
		status.Phase = onecloudv1.ResourceFailed
	case "ready":
		status.Phase = onecloudv1.ResourceFinished
	default:
		status.Phase = onecloudv1.ResourcePending
	}

	status.Reason = fmt.Sprintf("Exec '%s' successfully", extInfo.Action)

	return status, nil
}
