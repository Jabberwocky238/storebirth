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
func CreateUser(uid, email, passwordHash, secretKey string) (string, error) {
	var userUID string
	err := DB.QueryRow(
		"INSERT INTO users (uid, email, password_hash, secret_key) VALUES ($1, $2, $3, $4) RETURNING uid",
		uid, email, passwordHash, secretKey,
	).Scan(&userUID)
	return userUID, err
}

// GetUserByEmail 通过邮箱获取用户
func GetUserByEmail(email string) (*User, error) {
	var user User
	err := DB.QueryRow(
		"SELECT uid, email, password_hash, secret_key FROM users WHERE email = $1",
		email,
	).Scan(&user.UID, &user.Email, &user.PasswordHash, &user.SecretKey)
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

// GetUserSecretKey 通过 UID 获取用户密钥
func GetUserSecretKey(uid string) (string, error) {
	var secretKey string
	err := DB.QueryRow(
		"SELECT secret_key FROM users WHERE uid = $1",
		uid,
	).Scan(&secretKey)
	return secretKey, err
}

// ========== CustomDomain Actions ==========

// CreateCustomDomain 创建自定义域名
func CreateCustomDomain(id, userUID, domain, target, txtName, txtValue, status string) error {
	_, err := DB.Exec(
		`INSERT INTO custom_domains (id, user_uid, domain, target, txt_name, txt_value, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, userUID, domain, target, txtName, txtValue, status,
	)
	return err
}

// GetCustomDomain 获取自定义域名
func GetCustomDomain(id string) (*CustomDomain, error) {
	var cd CustomDomain
	err := DB.QueryRow(
		`SELECT id, user_uid, domain, target, txt_name, txt_value, status, created_at
		 FROM custom_domains WHERE id = $1`,
		id,
	).Scan(&cd.ID, &cd.UserUID, &cd.Domain, &cd.Target, &cd.TXTName, &cd.TXTValue, &cd.Status, &cd.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &cd, nil
}

// ListCustomDomains 获取用户的所有自定义域名
func ListCustomDomains(userUID string) ([]*CustomDomain, error) {
	rows, err := DB.Query(
		`SELECT id, user_uid, domain, target, txt_name, txt_value, status, created_at
		 FROM custom_domains WHERE user_uid = $1`,
		userUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*CustomDomain
	for rows.Next() {
		var cd CustomDomain
		if err := rows.Scan(&cd.ID, &cd.UserUID, &cd.Domain, &cd.Target, &cd.TXTName, &cd.TXTValue, &cd.Status, &cd.CreatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, &cd)
	}
	return domains, nil
}

// UpdateCustomDomainStatus 更新自定义域名状态
func UpdateCustomDomainStatus(id, status string) error {
	_, err := DB.Exec(
		`UPDATE custom_domains SET status = $1 WHERE id = $2`,
		status, id,
	)
	return err
}

// DeleteCustomDomain 删除自定义域名
func DeleteCustomDomain(id string) error {
	_, err := DB.Exec(`DELETE FROM custom_domains WHERE id = $1`, id)
	return err
}

// ========== Worker Actions ==========

// CreateWorker 创建 worker 记录
func CreateWorker(userUID, workerID, workerName string) error {
	_, err := DB.Exec(
		`INSERT INTO workers (user_uid, worker_id, worker_name) VALUES ($1, $2, $3)`,
		userUID, workerID, workerName,
	)
	return err
}

// GetWorkerByID 通过 worker_id 获取 worker
func GetWorkerByID(workerID string) (*Worker, error) {
	var w Worker
	err := DB.QueryRow(
		`SELECT user_uid, worker_id, worker_name, active_version_id, created_at
		 FROM workers WHERE worker_id = $1`, workerID,
	).Scan(&w.UserUID, &w.WorkerID, &w.WorkerName, &w.ActiveVersionID, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// ListWorkersByUser 获取用户的所有 worker
func ListWorkersByUser(userUID string) ([]*Worker, error) {
	rows, err := DB.Query(
		`SELECT user_uid, worker_id, worker_name, active_version_id, created_at
		 FROM workers WHERE user_uid = $1 ORDER BY created_at DESC`, userUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workers []*Worker
	for rows.Next() {
		var w Worker
		if err := rows.Scan(&w.UserUID, &w.WorkerID, &w.WorkerName, &w.ActiveVersionID, &w.CreatedAt); err != nil {
			return nil, err
		}
		workers = append(workers, &w)
	}
	return workers, nil
}

// ========== WorkerDeployVersion Actions ==========

// CreateDeployVersion 创建部署版本，返回 version id
func CreateDeployVersion(workerID, image string, port int) (int, error) {
	var id int
	err := DB.QueryRow(
		`INSERT INTO worker_deploy_versions (worker_id, image, port, status)
		 VALUES ($1, $2, $3, 'loading') RETURNING id`,
		workerID, image, port,
	).Scan(&id)
	return id, err
}

// UpdateDeployVersionStatus 更新部署版本状态和消息
func UpdateDeployVersionStatus(versionID int, status, msg string) error {
	_, err := DB.Exec(
		`UPDATE worker_deploy_versions SET status = $1, msg = $2 WHERE id = $3`,
		status, msg, versionID,
	)
	return err
}

// GetDeployVersion 获取部署版本
func GetDeployVersion(versionID int) (*WorkerDeployVersion, error) {
	var v WorkerDeployVersion
	err := DB.QueryRow(
		`SELECT id, worker_id, image, port, status, msg, created_at
		 FROM worker_deploy_versions WHERE id = $1`, versionID,
	).Scan(&v.ID, &v.WorkerID, &v.Image, &v.Port, &v.Status, &v.Msg, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// ListDeployVersions 获取 worker 的部署版本，支持分页
func ListDeployVersions(workerID string, limit, offset int) ([]*WorkerDeployVersion, error) {
	rows, err := DB.Query(
		`SELECT id, worker_id, image, port, status, msg, created_at
		 FROM worker_deploy_versions WHERE worker_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		workerID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*WorkerDeployVersion
	for rows.Next() {
		var v WorkerDeployVersion
		if err := rows.Scan(&v.ID, &v.WorkerID, &v.Image, &v.Port, &v.Status, &v.Msg, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, &v)
	}
	return versions, nil
}

// SetWorkerActiveVersion 设置 worker 的 active_version_id
func SetWorkerActiveVersion(workerID string, versionID int) error {
	_, err := DB.Exec(
		`UPDATE workers SET active_version_id = $1 WHERE worker_id = $2`,
		versionID, workerID,
	)
	return err
}

// DeleteWorker 删除 worker 记录
func DeleteWorker(workerID string) error {
	_, err := DB.Exec(`DELETE FROM workers WHERE worker_id = $1`, workerID)
	return err
}

// ListAllSuccessDomains 获取所有成功状态的域名（用于定期检查）
func ListAllSuccessDomains() ([]*CustomDomain, error) {
	rows, err := DB.Query(
		`SELECT id, user_uid, domain, target, txt_name, txt_value, status, created_at
		 FROM custom_domains WHERE status = 'success'`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*CustomDomain
	for rows.Next() {
		var cd CustomDomain
		if err := rows.Scan(&cd.ID, &cd.UserUID, &cd.Domain, &cd.Target, &cd.TXTName, &cd.TXTValue, &cd.Status, &cd.CreatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, &cd)
	}
	return domains, nil
}
