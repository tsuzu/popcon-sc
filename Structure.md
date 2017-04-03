# 構成

### ppweb
- Webサーバ部
- ユーザからのアクセスを処理する

### ppjc
- ジャッジコントローラ
- 各ジャッジノードに対して処理する命令を出す
- ランキングを更新
- MySQLのSubmissionテーブルを更新

### ppms
- その他雑用処理
- 不要になったファイルをGridFSから遅延削除

### MySQL
- 多くのデータを保存しておくためのデータベース
- 以下テーブル一覧
    - contest_participations: Swarm移行で使わなくなった
    - contest_problem_score_sets: 問題のスコアセットを保持
    - contest_problem_test_cases: 問題のテストケースデータの名前とファイルパスを保持
    - contest_problems: 問題の名前や制限など
    - contests: コンテストの名前や開催時間など
    - groups: ユーザが所属するグループの処理。0は入っていないがadmin
    - languages: ジャッジに使われる言語
    - news: トップページに表示されるNews
    - sessions: セッション情報
    - submission_test_cases: 各提出の各テストケースのジャッジ結果を保持
    - submissions: 各提出の結果やファイルパスの保存
    - users: ユーザのIDやパスワードの保持

### Redis
- CSRF対策Tokenやメール認証のトークン及びセッション
- また設定も多くを保存している

### MongoDB
- GridFSのみ使用
- 各部のファイルを全て保存している

## 今後増える予定
### pplb
- ロードバランサ(traefikを使うかも)
- MongoDB/GridFSはSeaweedFSになるかも
- ジャッジノード