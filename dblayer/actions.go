package dblayer

import "time"

// ========== User Actions ==========

// GetVerificationCode 获取验证码
func GetVerificationCode(email, code string) (int, time.Time, error) {
	var codeID int
	var expiresAt time.Time
	err := DB.QueryRow(
		"SELECT id, expires_at FROM verification_codes WHERE email = $1 AND code = $2 AND used = false",
		email, code,
	).Scan(&codeID, &expiresAt)
	return codeID, expiresAt, err
}

// MarkCodeUsed 标记验证码已使用
func MarkCodeUsed(codeID int) error {
	_, err := DB.Exec("UPDATE verification_codes SET used = true WHERE id = $1", codeID)
	return err
}

// CreateUser 创建用户
func CreateUser(uid, email, passwordHash string) (string, error) {
	var userUID string
	err := DB.QueryRow(
		"INSERT INTO users (uid, email, password_hash) VALUES ($1, $2, $3) RETURNING uid",
		uid, email, passwordHash,
	).Scan(&userUID)
	return userUID, err
}

// GetUserByEmail 通过邮箱获取用户
func GetUserByEmail(email string) (*User, error) {
	var user User
	err := DB.QueryRow(
		"SELECT uid, email, password_hash FROM users WHERE email = $1",
		email,
	).Scan(&user.UID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SaveVerificationCode 保存验证码
func SaveVerificationCode(email, code string, expiresAt time.Time) error {
	_, err := DB.Exec(
		"INSERT INTO verification_codes (email, code, expires_at) VALUES ($1, $2, $3)",
		email, code, expiresAt,
	)
	return err
}

// UpdateUserPassword 更新用户密码
func UpdateUserPassword(email, passwordHash string) error {
	_, err := DB.Exec(
		"UPDATE users SET password_hash = $1 WHERE email = $2",
		passwordHash, email,
	)
	return err
}

// ========== RDB Actions ==========

// CreateRDB 创建 RDB 资源
func CreateRDB(userUID, rdbUID, name, rdbType, url string) (string, error) {
	var uid string
	err := DB.QueryRow(
		`INSERT INTO user_rdbs (user_id, uid, name, rdb_type, url)
		 VALUES ((SELECT id FROM users WHERE uid = $1), $2, $3, $4, $5)
		 RETURNING uid`,
		userUID, rdbUID, name, rdbType, url,
	).Scan(&uid)
	return uid, err
}

// ListRDBsByUser 获取用户的所有 RDB
func ListRDBsByUser(userUID string) ([]RDB, error) {
	rows, err := DB.Query(
		`SELECT uid, name, rdb_type, url, enabled FROM user_rdbs
		 WHERE user_id = (SELECT id FROM users WHERE uid = $1)`,
		userUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rdbs []RDB
	for rows.Next() {
		var r RDB
		rows.Scan(&r.UID, &r.Name, &r.Type, &r.URL, &r.Enabled)
		rdbs = append(rdbs, r)
	}
	return rdbs, nil
}

// DeleteRDB 删除 RDB 资源
func DeleteRDB(rdbUID, userUID string) (int64, error) {
	result, err := DB.Exec(
		`DELETE FROM user_rdbs
		 WHERE uid = $1 AND user_id = (SELECT id FROM users WHERE uid = $2)`,
		rdbUID, userUID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ========== KV Actions ==========

// CreateKV 创建 KV 资源
func CreateKV(userUID, kvUID, name, kvType, url string) (string, error) {
	var uid string
	err := DB.QueryRow(
		`INSERT INTO user_kvs (user_id, uid, name, kv_type, url)
		 VALUES ((SELECT id FROM users WHERE uid = $1), $2, $3, $4, $5)
		 RETURNING uid`,
		userUID, kvUID, name, kvType, url,
	).Scan(&uid)
	return uid, err
}

// ListKVsByUser 获取用户的所有 KV
func ListKVsByUser(userUID string) ([]KV, error) {
	rows, err := DB.Query(
		`SELECT uid, name, kv_type, url, enabled FROM user_kvs
		 WHERE user_id = (SELECT id FROM users WHERE uid = $1)`,
		userUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var kvs []KV
	for rows.Next() {
		var k KV
		rows.Scan(&k.UID, &k.Name, &k.Type, &k.URL, &k.Enabled)
		kvs = append(kvs, k)
	}
	return kvs, nil
}

// DeleteKV 删除 KV 资源
func DeleteKV(kvUID, userUID string) (int64, error) {
	result, err := DB.Exec(
		`DELETE FROM user_kvs
		 WHERE uid = $1 AND user_id = (SELECT id FROM users WHERE uid = $2)`,
		kvUID, userUID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ========== Task Actions ==========

// EnqueueConfigTask 添加配置更新任务
func EnqueueConfigTask(userUID string) (int, error) {
	var taskID int
	err := DB.QueryRow(`
		INSERT INTO config_tasks (user_uid, task_type, status)
		VALUES ($1, 'config_update', 'pending')
		RETURNING id
	`, userUID).Scan(&taskID)
	return taskID, err
}

// EnqueuePodCreateTask 添加 Pod 创建任务
func EnqueuePodCreateTask(userUID string) (int, error) {
	var taskID int
	err := DB.QueryRow(`
		INSERT INTO config_tasks (user_uid, task_type, status)
		VALUES ($1, 'pod_create', 'pending')
		RETURNING id
	`, userUID).Scan(&taskID)
	return taskID, err
}

// GetTaskStatus 获取任务状态
func GetTaskStatus(taskID int) (string, string, error) {
	var status, errorMsg string
	err := DB.QueryRow(`
		SELECT status, COALESCE(error_msg, '') FROM config_tasks WHERE id = $1
	`, taskID).Scan(&status, &errorMsg)
	return status, errorMsg, err
}

// FetchPendingTask 获取并锁定一个待处理任务
func FetchPendingTask() (int, string, string, error) {
	var taskID int
	var userUID, taskType string
	err := DB.QueryRow(`
		UPDATE config_tasks
		SET status = 'processing', updated_at = NOW()
		WHERE id = (
			SELECT id FROM config_tasks
			WHERE status = 'pending'
			ORDER BY created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, user_uid, task_type
	`).Scan(&taskID, &userUID, &taskType)
	return taskID, userUID, taskType, err
}

// MarkTaskFailed 标记任务失败
func MarkTaskFailed(taskID int, errMsg string) error {
	_, err := DB.Exec(`
		UPDATE config_tasks
		SET status = 'failed', error_msg = $1, updated_at = NOW()
		WHERE id = $2
	`, errMsg, taskID)
	return err
}

// MarkTaskCompleted 标记任务完成
func MarkTaskCompleted(taskID int) error {
	_, err := DB.Exec(`
		UPDATE config_tasks
		SET status = 'completed', updated_at = NOW()
		WHERE id = $1
	`, taskID)
	return err
}

// ========== Config Generation ==========

// RDBConfigItem RDB 配置项
type RDBConfigItem struct {
	UID  string
	Type string
	URL  string
}

// KVConfigItem KV 配置项
type KVConfigItem struct {
	UID  string
	Type string
	URL  string
}

// GetUserRDBsForConfig 获取用户启用的 RDB 配置
func GetUserRDBsForConfig(userUID string) ([]RDBConfigItem, error) {
	rows, err := DB.Query(
		`SELECT uid, rdb_type, url FROM user_rdbs
		 WHERE user_id = (SELECT id FROM users WHERE uid = $1) AND enabled = true`,
		userUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []RDBConfigItem
	for rows.Next() {
		var item RDBConfigItem
		rows.Scan(&item.UID, &item.Type, &item.URL)
		items = append(items, item)
	}
	return items, nil
}

// GetUserKVsForConfig 获取用户启用的 KV 配置
func GetUserKVsForConfig(userUID string) ([]KVConfigItem, error) {
	rows, err := DB.Query(
		`SELECT uid, kv_type, url FROM user_kvs
		 WHERE user_id = (SELECT id FROM users WHERE uid = $1) AND enabled = true`,
		userUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []KVConfigItem
	for rows.Next() {
		var item KVConfigItem
		rows.Scan(&item.UID, &item.Type, &item.URL)
		items = append(items, item)
	}
	return items, nil
}
