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

type WrapperRedPacket struct {
	*WrapperCommon
}

func NewWrapperRedPacket(wrapperCommon *WrapperCommon) *WrapperRedPacket {
	return &WrapperRedPacket{WrapperCommon: wrapperCommon}
}

func (w *WrapperRedPacket) RedPacketCreateOrder(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.RedPacketCreateOrder, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperRedPacket) RedPacketCreatedCallback(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.RedPacketCreatedCallback, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperRedPacket) RedPacketGetDetail(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.RedPacketGetDetail, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperRedPacket) RedPacketIssueClaimSign(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.RedPacketIssueClaimSign, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperRedPacket) RedPacketClaimResult(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.RedPacketClaimResult, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperRedPacket) RedPacketRequestRefund(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.RedPacketRequestRefund, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperRedPacket) RedPacketGetRefund(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.RedPacketGetRefund, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperRedPacket) RedPacketIssueWalletBindChallenge(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.RedPacketIssueWalletBindChallenge, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperRedPacket) RedPacketConfirmWalletBind(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.RedPacketConfirmWalletBind, callback, &args).AsyncCallWithCallback()
}

func (w *WrapperRedPacket) RedPacketGetWalletBinding(_ js.Value, args []js.Value) interface{} {
	callback := event_listener.NewBaseCallback(utils.FirstLower(utils.GetSelfFuncName()), w.commonFunc)
	return event_listener.NewCaller(open_im_sdk.RedPacketGetWalletBinding, callback, &args).AsyncCallWithCallback()
}
