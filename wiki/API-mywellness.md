# mywellness API (upstream)

Headers on every call: `X-MWAPPS-APPID: EC1D38D7-D359-48D0-A60C-D8C0B8FB9DF9`, `X-MWAPPS-CLIENT: enduserweb`, `X-MWAPPS-CLIENTVERSION`, query `_c=it-IT`; after login `Authorization: Bearer <token>`.

| Purpose | Endpoint | Notes |
|---|---|---|
| Login | `POST https://core.mywellness.com/v2/enduser/authentication/login` `{username,password,keepMeLoggedIn}` | → `token`, `userContext.id`; `MfaRequired` (401) not supported |
| Session check | `POST https://services.mywellness.com/application/<appId>/GetLoginStatus` | 200 with `errors[].field = TokenNotValid` when expired (tokens last ~1 h; other APIs then answer as anonymous) |
| Club | `GET https://core.mywellness.com/v2/enduser/facility/detail?facilityUrl=…` | `id`, `name` |
| Schedule | `GET https://calendar.mywellness.com/v2/enduser/class/Search?eventTypes=Class&facilityId=…&fromDate=YYYY-MM-DD&toDate=…` | public; with token `isParticipant`, `isInWaitingList`, `waitingListPosition` |
| Book | `POST https://calendar.mywellness.com/v2/enduser/class/Book` `{partitionDate:YYYYMMDD,userId,classId,station:null}` | `result` ∈ `Booked`, `UserAddedToWaitingList`, `PlaceNotAvailable`, `ToMuchParticipants`, `Failed` |
| Unbook | `POST https://calendar.mywellness.com/v2/enduser/class/Unbook` | same body; `result` ∈ `UnBooked`, `TooLate`, `BookingNotAvailable`, `EventNotExists`, `UserNotExists`, `Failed` |
| Leave waiting list | `POST https://services.mywellness.com/core/calendarevent/{classId}/RemoveFromWaitingList` `{partitionDate:"YYYYMMDD",userId}` | `data` ∈ `Removed`, `UserNotInWaitingList`, `Failed` |

Useful class fields: `id`, `partitionDate`, `startDate`, `name`, `room`, `assignedTo`, `pictureUrl`, `maxParticipants`, `numberOfParticipants`, `availablePlaces`, `isParticipant`, `isInWaitingList`, `bookingInfo.bookingOpensOn`, `bookingInfo.bookingAvailable`, `bookingInfo.cancellationMinutesInAdvance` (120), `bookingInfo.bookingCloseMinutesInAdvance` (5), `bookingInfo.bookingHasWaitingList`.

Note: the mywellness waiting list only notifies; it does not book automatically. That is why the gateway keeps watching.
