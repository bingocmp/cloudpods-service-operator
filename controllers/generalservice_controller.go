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

package controllers

import (
	"context"
	"fmt"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
	onecloudv1 "yunion.io/x/onecloud-service-operator/api/v1"
	"yunion.io/x/onecloud-service-operator/pkg/resources"
)

// GeneralServiceReconciler reconciles a GeneralService object
type GeneralServiceReconciler struct {
	ReconcilerBase
	// Enable intensive information collection during the reconcile process
	Dense bool
}

// +kubebuilder:rbac:groups=onecloud.yunion.io,resources=generalservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=onecloud.yunion.io,resources=generalservices/status,verbs=get;update;patch

func (r *GeneralServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	var generalService = &onecloudv1.GeneralService{}
	if err := r.Get(ctx, req.NamespacedName, generalService); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log := r.GetLog(generalService)
	remoteGS := resources.NewGeneralService(generalService, log)

	dealErr := func(err error) (ctrl.Result, error) {
		return r.dealErr(ctx, remoteGS, err)
	}

	has, ret, err := r.UseFinallizer(ctx, remoteGS)
	if !has {
		return ret, err
	}

	params := make(map[string]interface{}, len(generalService.Spec.Params))
	for varName, sv := range generalService.Spec.Params {
		vv, err := sv.GetValue(ctx)
		if err != nil {
			generalService.GetResourceStatus().SetPhase(onecloudv1.ResourceInvalid,
				fmt.Sprintf("The value of var '%s' is valid: %s", varName, err.Error()),
			)
			return ctrl.Result{}, r.Status().Update(ctx, generalService)
		}
		if vv == nil {
			continue
		}
		params[varName] = vv.Interface()
	}

	if generalService.Status.ExternalInfo.Id == "" {
		log.Info("Service create", "params", params)

		return r.Create(ctx, remoteGS, params, true)
	}

	if generalService.Status.Phase == onecloudv1.ResourceInvalid {
		return ctrl.Result{}, nil
	}

	if generalService.Status.Phase == onecloudv1.ResourcePending {
		rs, err := remoteGS.GetStatus(ctx)
		if err != nil {
			return dealErr(err)
		}
		if r.requireUpdate(generalService, rs.(*onecloudv1.GeneralServiceStatus)) {
			generalService.SetResourceStatus(rs)
			return ctrl.Result{}, r.Status().Update(ctx, generalService)
		}
		return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}

	detail, err := remoteGS.GetDetail(ctx)
	if err != nil {
		return dealErr(err)
	}

	changed := false
	if reqParams, exist := detail["request_params"].(map[string]interface{}); exist {
		for k := range params {
			if fmt.Sprintf("%v", reqParams[k]) != fmt.Sprintf("%v", params[k]) {
				changed = true
				break
			}
		}
	}
	if !changed {
		return ctrl.Result{}, nil
	}

	log.Info("Service update", "params", params)

	status, err := remoteGS.Update(ctx, params)
	if err != nil {
		return dealErr(err)
	}
	if r.requireUpdate(generalService, status) {
		generalService.SetResourceStatus(status)
		return ctrl.Result{}, r.Status().Update(ctx, generalService)
	}

	return ctrl.Result{}, nil
}

func (r *GeneralServiceReconciler) requireUpdate(gs *onecloudv1.GeneralService, newStatus *onecloudv1.GeneralServiceStatus) bool {
	if newStatus == nil {
		return false
	}
	if gs.Status.Phase != newStatus.Phase || gs.Status.Reason != newStatus.Reason {
		return true
	}
	return false
}

func (r *GeneralServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	gs := &onecloudv1.GeneralService{}
	return ctrl.NewControllerManagedBy(mgr).
		For(gs).
		Watches(
			&source.Kind{Type: &onecloudv1.GeneralService{}},
			&handler.EnqueueRequestForOwner{
				OwnerType:    gs,
				IsController: false,
			}).
		Complete(r)
}
