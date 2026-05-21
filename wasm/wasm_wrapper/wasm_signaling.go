// Copyright © 2023 OpenIM SDK. All rights reserved.
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

//go:build js && wasm
// +build js,wasm

package wasm_wrapper

import (
	"syscall/js"

	"github.com/openimsdk/openim-sdk-core/v3/open_im_sdk"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/utils"
	"github.com/openimsdk/openim-sdk-core/v3/wasm/event_listener"
)

type WrapperSignaling struct {
	*WrapperCommon
}

func NewWrapperSignaling(wrapperCommon *WrapperCommon) *WrapperSignaling {
	return &WrapperSignaling{WrapperCommon: wrapperCommon}
}

func (w *WrapperSignaling) SignalingInviteInGroup(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingInviteInGroup, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperSignaling) SignalingInvite(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingInvite, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperSignaling) SignalingAccept(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingAccept, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperSignaling) SignalingReject(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingReject, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperSignaling) SignalingCancel(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingCancel, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperSignaling) SignalingTimeout(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingTimeout, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperSignaling) SignalingHungUp(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingHungUp, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperSignaling) SignalingGetRoomByGroupID(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingGetRoomByGroupID, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperSignaling) SignalingGetTokenByRoomID(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingGetTokenByRoomID, callback, &args).AsyncCallWithCallback()
}

// SignalingGetInvitationRecords 分页查询音视频通话记录（对接 RTC GetSignalInvitationRecords）。
func (w *WrapperSignaling) SignalingGetInvitationRecords(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingGetInvitationRecords, callback, &args).AsyncCallWithCallback()
}

// SignalingGetLocalCallRecords 按用户查询本地通话记录（status：0=全部 1=已接听 2=未接通）。
func (w *WrapperSignaling) SignalingGetLocalCallRecords(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingGetLocalCallRecords, callback, &args).AsyncCallWithCallback()
}

// SignalingSearchLocalCallRecords 查询 SDK 本地缓存的通话记录表。
func (w *WrapperSignaling) SignalingSearchLocalCallRecords(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingSearchLocalCallRecords, callback, &args).AsyncCallWithCallback()
}

// SignalingGetLocalCallRecordDetail 按 sID 查询单条本地通话详情。
func (w *WrapperSignaling) SignalingGetLocalCallRecordDetail(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingGetLocalCallRecordDetail, callback, &args).AsyncCallWithCallback()
}

// SignalingDeleteLocalCallRecords 删除本地音视频通话记录（支持批量 sID）。
func (w *WrapperSignaling) SignalingDeleteLocalCallRecords(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingDeleteLocalCallRecords, callback, &args).AsyncCallWithCallback()
}

// SignalingDeleteSignalRecords 删除服务端 + 本地通话记录。
func (w *WrapperSignaling) SignalingDeleteSignalRecords(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingDeleteSignalRecords, callback, &args).AsyncCallWithCallback()
}

// SignalingClearAllLocalCallRecords 清空本地所有音视频通话记录。
func (w *WrapperSignaling) SignalingClearAllLocalCallRecords(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingClearAllLocalCallRecords, callback, &args).AsyncCallWithCallback()
}

// SignalingClearAllCallRecords 清空服务端 + 本地所有记录（当前主要清本地）。
func (w *WrapperSignaling) SignalingClearAllCallRecords(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingClearAllCallRecords, callback, &args).AsyncCallWithCallback()
}

// SignalingSendCustomSignal 发送自定义信令。
func (w *WrapperSignaling) SignalingSendCustomSignal(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.SignalingSendCustomSignal, callback, &args).AsyncCallWithCallback()
}
