# GOでBGPを実装
"作って学ぶルーティングプロトコル　RustでBGPを実装"を参考に、Go言語でBGPを実装していく。
- https://nextpublishing.jp/book/15905.html
- https://github.com/Miyoshi-Ryota/how-to-create-bgp
- https://github.com/Miyoshi-Ryota/mrbgpdv2

## 実装の段階
以下の段階に分けてイベント駆動ステートマシンとしてBGPを実装していく。
なお、正常系のみ実装する。
1. "Connect"まで遷移する実装
2. "Established"まで遷移する実装
3. "Update Message"を交換する実装

## BGPのイベント駆動ステートマシンに登場するEvent   (本書から抜粋)
|Event名|説明|
|---|---|
|ManualStart|BGPの開始を指示したときに発行されるイベント。<br>RFCないでも同様に定義されている。|
|TcpConnectionConfirmed|対向機器とTCPコネクションを確率できたときに発行されるイベント。<br>RFC内でも同様に定義されている。<br>RFCでは、TCP ackを受信したときのイベント、TcpCrAckedと区別している。<br>しかし本書では正常系しか実装しないため、TCPコネクションが確立された時のイベントとしては本イベントのみにしている。|
|BGPOpen|対向機器からOpen Massageを受信したときに発行されるイベント。<br>RFC内でも同様に定義されている。|
|KeepAliveMsg|対向機器からKeepAlive Massageを受信したときに発行されるイベント。<br>RFC内でも同様に定義されている。|
|UpdateMsg|対向機器からUpdate Messageを受信したときに発行されるイベント。<br>RFC内でも同様に定義されている。|
|Established|Established Stateに遷移したときに発行されるイベント。<br>存在するほうが実装しやすいため追加した。<br>RFCには存在しないイベント。|
|LocRibChanged|LocRibが変更がされたときに発行されるイベント。<br>存在するほうが実装しやすいため追加した。<br>RFCには存在しないイベント。|
|AdjRibInChanged|AdjRibInが変更がされたときに発行されるイベント。<br>存在するほうが実装しやすいため追加した。<br>RFCには存在しないイベント。|
|AdjRibOutChanged|AdjRibOutが変更がされたときに発行されるイベント。<br>存在するほうが実装しやすいため追加した。<br>RFCには存在しないイベント。|

## BGPのイベント駆動ステートマシンに登場するState   (本書から抜粋)
|State名|説明|
|---|---|
|Idle|初期状態|
|Connect|TCPコネクションの確立を待機している状態|
|OpenSent|PeerからのOpen Messageを待機している状態|
|OpenConfirm|PeerからのKeepAlive Messageを待機している状態|
|Established|Peerが正常に確立され、Update Messageなどのやり取りが可能になった状態|

## BGPのステートマシン
```mermaid
stateDiagram-v2
  [*] --> Idle
  Idle --> Connect:<b>ManualStart Event</b><br>対向機器とTCPコネクションの作成を試みる
  Connect --> OpenSent:<b>TcpConnectionConfirmed Event</b><br>対向機器にBGP Open Messageを送信する
  OpenSent --> OpenConfirm:<b>BGPOpen Event</b><br>対向機器にBGP Keepalive Messageを送信する
  OpenConfirm --> Established:<b>KeepAliveMsg Event</b>
  EstablishedEvent --> Established:LocRibの内容をAdjRibOutに反映する
  UpdateMsgEvent --> Established:受信したBGP Update Messageの内容から AdjRibInを更新する
  AdjRibInChangedEvent --> Established:AdjRibInの情報からLocRibを更新する
  LocRibChangedEvent --> Established:LocRibの内容をAdjRibOutに反映する
  AdjRibOutChangedEvent --> Established:Update Messageを作成し、対向機器へ送信する
  Established --> EstablishedEvent
  Established --> UpdateMsgEvent
  Established --> AdjRibInChangedEvent
  Established --> LocRibChangedEvent
  Established --> AdjRibOutChangedEvent

  style EstablishedEvent fill:#eee,stroke:#fff
  style UpdateMsgEvent fill:#eee,stroke:#fff
  style AdjRibInChangedEvent fill:#eee,stroke:#fff
  style LocRibChangedEvent fill:#eee,stroke:#fff
  style AdjRibOutChangedEvent fill:#eee,stroke:#fff
  ```

## BGP Message
### Headerフォーマット(19byte)
|名前|bit数|Octet|説明|
|---|---|---|---|
|Marker|128|1-16|全て1。互換性のために存在|
|Length|16|17-18|Headerを含めたBGP Message全体のバイト数を表す符号なし整数値|
|Type|8|19|BGP Messageの種類を表す符号なし整数値。<br>1: OPEN<br>2: UPDATE<br>3: NOTIFICATION<br>4: KEEPALIVE|

### Open Messageフォーマット(29 + option byte)
|名前|bit数|Octet|説明|
|---|---|---|---|
|Header|152|1-19|BGP Message Header|
|Version|8|20|BGPのバージョンを表す符号なし整数値<br>現在のVersionは4|
|My Autonomous System|16|21-22|送信者のAS番号を表す符号なし整数値|
|Hold Time|16|23-24|Hold Timerの秒数を表す符号なし整数値<br>BGPではEstablishedになった後、<br>定期的にKeepalive Messageを交換する<br>HoldTimeの秒数だけKeepaliveを受信できなかった時、<br>Peerがダウンしていると見なす<br>0でこの機能を使用しないことを表す<br>本実装ではHold Timerを実装しないため0とする #TODO|
|BGP Identifer|32|25-28|送信者のIPアドレス(?)|
|Optional Parameters Length|8|29|Optional Parametersのオクテット数を表す符号なし整数値|
|Optional Parameters|非固定||オプショナルなパラメータ<br>本実装では使用しない|