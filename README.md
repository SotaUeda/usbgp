# GOでBGPを実装
"作って学ぶルーティングプロトコル　RustでBGPを実装"を参考に、Go言語でBGPを実装していく。
- https://nextpublishing.jp/book/15905.html
- https://github.com/Miyoshi-Ryota/how-to-create-bgp
- https://github.com/Miyoshi-Ryota/mrbgpdv2

## TODO
- Marshalメソッドはerrorを返す必要がないかも
  - オブジェクトが作成する段階でエラー処理、あとはイミュータブルにしておく
  - UnMarshalは受信したデータが不正の可能性があるからerrorが必須

## 実装の段階
以下の段階に分けてイベント駆動ステートマシンとしてBGPを実装していく。
なお、正常系のみ実装する。
1. "Connect"まで遷移する実装
2. "Established"まで遷移する実装
3. "Update Message"を交換する実装

## BGPのイベント駆動ステートマシンに登場するEvent   (本書から抜粋)
|Event名|説明|
|---|---|
|ManualStart|BGPの開始を指示したときに発行されるイベント<br>RFCないでも同様に定義されている。|
|TcpConnectionConfirmed|対向機器とTCPコネクションを確率できたときに発行されるイベント<br>RFC内でも同様に定義されている<br>RFCでは、TCP ackを受信したときのイベント、TcpCrAckedと区別している<br>しかし本書では正常系しか実装しないため、TCPコネクションが確立された時のイベントとしては本イベントのみにしている。|
|BGPOpen|対向機器からOpen Massageを受信したときに発行されるイベント<br>RFC内でも同様に定義されている。|
|KeepAliveMsg|対向機器からKeepAlive Massageを受信したときに発行されるイベント<br>RFC内でも同様に定義されている。|
|UpdateMsg|対向機器からUpdate Messageを受信したときに発行されるイベント<br>RFC内でも同様に定義されている。|
|Established|Established Stateに遷移したときに発行されるイベント<br>存在するほうが実装しやすいため追加した<br>RFCには存在しないイベント。|
|LocRibChanged|LocRibが変更がされたときに発行されるイベント<br>存在するほうが実装しやすいため追加した<br>RFCには存在しないイベント。|
|AdjRibInChanged|AdjRibInが変更がされたときに発行されるイベント<br>存在するほうが実装しやすいため追加した<br>RFCには存在しないイベント。|
|AdjRibOutChanged|AdjRibOutが変更がされたときに発行されるイベント<br>存在するほうが実装しやすいため追加した<br>RFCには存在しないイベント。|

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
|Type|8|19|BGP Messageの種類を表す符号なし整数値<br>1: OPEN<br>2: UPDATE<br>3: NOTIFICATION<br>4: KEEPALIVE|

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

### Keepalive Messageフォーマット(19byte Headerのみ)
|名前|bit数|Octet|説明|
|---|---|---|---|
|Header|152|1-19|BGP Message Header|

### Update Messageフォーマット(23 + 非固定 byte)
|名前|bit数|説明|
|---|---|---|
|Header|152|BGP Message Header|
|Withdrawn Routes Length|16|Withdrawn Routesのオクテット数を表す符号なし整数値|
|Withdrawn Routes|非固定|使用できなくなった宛先情報。複数の宛先情報を同時に持つ<br>1つのRouteはPrefix長とネットワークアドレスからなる<br>バイト列の表現では、Prefix長は1オクテットの符号なし整数値<br>ネットワークアドレスは、可変長で最短のオクテット数で表す<br>具体例として192.168.0.0/16の経路をバイト列で表すと[16.192.168]の3オクテットのデータになる。|
|Total Path Attribute Length|16|Path Attributesのオクテット数を表す符号なし整数値|
Path Attributes|非固定|経路属性情報。経路選択の計算に使用される付加情報<br>複数のPathAttributeを同時に持つ<br>1つのPathAttributeは、Attribute Type, Attribute Lentgth, Attribute Valueの3つの情報から構成される<br>Attribute TypeはAttr Flags, Attr Type Codeの2つの情報から構成される<br>Attr Flagsはさらに細切れの情報からなる。詳細は別表|
Network Layer Reachability Information|非固定|使用可能な宛先情報<br>バイト列表現はWithdrawn Routesと同様|

### PathAttributeのフォーマット
|名前|bit数|説明|
|---|---|---|
|Optional bit|1|Path Attributeには全てのBGP実装が実装するべきであるWell-knownと、実装が任意であるOptionalの二種類が存在する<br>OptionalなPath Attributeの場合、本bitを1にする<br>Well-knownなPath Attributeの場合、本bitを0にする|
|Transitive bit|1|別のネイバーにも経路を送信する際にも、このPath Attributeを必ず保持・通知する場合は本bitを1に、そうでない場合は0にする<br>なお、Well-knownなPathAttributeは必ず1にセットする|
|Partial bit|1|別のネイバーにも経路を送信する際にも、このPath Attributeを保持・通知する場合かどうか任意である場合は本bitを1に、そうでない場合は0にする。<br>なお、Well-knownなPathAttributeは必ず1にセットする|
|Extended Length bit|1|Attribute Lengthのオクテット数が1の場合は本bitを0にする<br>Attribute Lengthのオクテット数が2の場合は本bitを1にする|
|未使用のbit|4|用途はない。0にセットする|
|Attr Type Code|8|Path Attributeの種類を表す符号なし整数値<br>1でOrigin、<br>2でAS_Path、<br>3でNEXT_HOP<br>その他のコードが割り振られてるPath Attributeも存在するが、本実装ではこの3つのみ実装する|
|Attribute Length|非固定(8 or 16)|Attribute Valueのオクテット数を表す符号なし整数値|
|Attribute Value|非固定|Attr Type Codeによって表現が変わる<br>Originの場合、1オクテットのデータで、<br>0でこの経路をIGPで学習したことを、<br>1でEGPで学習したことを表す<br><br>AS_Pathの場合、Path Segment Type、Path Segment Length、Path Segment Valueの3つから構成される可変長のデータとなる<br>Path Segment Typeは1オクテットのデータで、<br>AS Pathを順序に意味のない集合(set)で扱う場合1に、<br>順序に意味のあるシーケンスとして扱う場合2にする<br>Path Segment Lengthは1オクテットのデータで、ASパスの数を表す整数である。<br>Path Segment Valueは可変長のデータを保持している<br>それぞれ1つのAS Pathは2オクテットずつのデータで表される|