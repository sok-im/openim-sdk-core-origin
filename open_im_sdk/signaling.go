package open_im_sdk

import "github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"

func SignalingInvite(callback open_im_sdk_callback.Base, operationID string, signalInviteReq string) {
	call(callback, operationID, UserForSDK.Signaling().Invite, signalInviteReq)
}

func SignalingInviteInGroup(callback open_im_sdk_callback.Base, operationID string, signalInviteInGroupReq string) {
	call(callback, operationID, UserForSDK.Signaling().InviteInGroup, signalInviteInGroupReq)
}

func SignalingAccept(callback open_im_sdk_callback.Base, operationID string, signalAcceptReq string) {
	call(callback, operationID, UserForSDK.Signaling().Accept, signalAcceptReq)
}

func SignalingReject(callback open_im_sdk_callback.Base, operationID string, signalRejectReq string) {
	call(callback, operationID, UserForSDK.Signaling().Reject, signalRejectReq)
}

func SignalingCancel(callback open_im_sdk_callback.Base, operationID string, signalCancelReq string) {
	call(callback, operationID, UserForSDK.Signaling().Cancel, signalCancelReq)
}

func SignalingHungUp(callback open_im_sdk_callback.Base, operationID string, signalHungUpReq string) {
	call(callback, operationID, UserForSDK.Signaling().HungUp, signalHungUpReq)
}

func SignalingGetTokenByRoomID(callback open_im_sdk_callback.Base, operationID string, signalGetTokenByRoomIDReq string) {
	call(callback, operationID, UserForSDK.Signaling().GetTokenByRoomID, signalGetTokenByRoomIDReq)
}

func SignalingGetRoomByGroupID(callback open_im_sdk_callback.Base, operationID string, groupID string) {
	call(callback, operationID, UserForSDK.Signaling().GetRoomByGroupID, groupID)
}

func SignalingGetInvitationInfoStartApp(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetSignalInvitationInfoStartApp, req)
}

func SignalingGetInvitationRecords(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetSignalInvitationRecords, req)
}

func SignalingSendCustomSignal(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().SendCustomSignal, req)
}
