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

func SignalingTimeout(callback open_im_sdk_callback.Base, operationID string, signalTimeoutReq string) {
	call(callback, operationID, UserForSDK.Signaling().Timeout, signalTimeoutReq)
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

func SignalingGetLocalCallRecords(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetLocalCallRecords, req)
}

func SignalingGetLocalCallRecordsWithUser(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetLocalCallRecordsWithUser, req)
}

func SignalingGetLocalMissedCallRecords(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetLocalMissedCallRecords, req)
}

func SignalingGetLocalAnsweredCallRecords(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetLocalAnsweredCallRecords, req)
}

func SignalingGetLocalAllCallRecords(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetLocalAllCallRecords, req)
}

func SignalingGetLocalCallRecordsByUserName(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().GetLocalCallRecordsByUserName, req)
}

func SignalingSearchLocalCallRecords(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().SearchLocalSignalCallRecords, req)
}

func SignalingGetLocalCallRecordDetail(callback open_im_sdk_callback.Base, operationID string, sID string) {
	call(callback, operationID, UserForSDK.Signaling().GetLocalSignalCallRecordDetail, sID)
}

func SignalingDeleteLocalCallRecords(callback open_im_sdk_callback.Base, operationID string, sIDs string) {
	call(callback, operationID, UserForSDK.Signaling().DeleteLocalSignalCallRecords, sIDs)
}

func SignalingDeleteSignalRecords(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().DeleteSignalRecords, req)
}

func SignalingClearAllLocalCallRecords(callback open_im_sdk_callback.Base, operationID string) {
	call(callback, operationID, UserForSDK.Signaling().ClearAllLocalSignalCallRecords)
}

func SignalingClearAllCallRecords(callback open_im_sdk_callback.Base, operationID string) {
	call(callback, operationID, UserForSDK.Signaling().ClearAllSignalCallRecords)
}

func SignalingSendCustomSignal(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Signaling().SendCustomSignal, req)
}
