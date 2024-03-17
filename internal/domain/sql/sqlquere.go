package sqlquere

var SyncUpdateTextTable = `WITH updated_rows AS (
	UPDATE text_data
	SET
	  name = $1,
	  data = $2,
	  uid = $3,
	  deleted = $4,
	  last_update = $5
	WHERE
	  name = $6
	  AND last_update < $7
	  AND uid = $8
	RETURNING *
  )
  INSERT INTO text_data (name, data, uid, deleted, last_update)
  SELECT
	$9,
	$10,
	$11,
	$12,
	$13
  WHERE
	NOT EXISTS (SELECT 1 FROM updated_rows)
	AND NOT EXISTS (SELECT *
	FROM text_data
	WHERE
	  name = $14
	  AND uid = $15)`

var SyncUpdateAuthTable = `WITH updated_rows AS (
		UPDATE logins
		SET
		  name = $1,
		  login = $2,
		  password = $3,
		  uId = $4,
		  deleted = $5,
		  last_update = $6
		WHERE
		  name = $7
		  AND last_update < $8
		  AND uid = $9 
		RETURNING *
	  )
	  INSERT INTO logins (name, login, password, uid, deleted, last_update)
	  SELECT
		$10,
		$11,
		$12,
		$13,
		$14,
		$15
	  WHERE
		NOT EXISTS (SELECT 1 FROM updated_rows)
		AND NOT EXISTS (SELECT *
		FROM logins
		WHERE
		  name = $16
		  AND uid = $17) `

var SyncUpdateBinTable = `WITH updated_rows AS (
			UPDATE binares_data
			SET
			  name = $1,
			  data = $2,
			  uId = $3,
			  deleted = $4,
			  last_update = $5
			WHERE
			  name = $6
			  AND last_update < $7
			  AND uid = $8
			RETURNING *
		  )
		  INSERT INTO binares_data (name, data, uid, deleted, last_update)
		  SELECT
			$9,
			$10,
			$11,
			$12,
			$13
		  WHERE
			NOT EXISTS (SELECT 1 FROM updated_rows)
			AND NOT EXISTS (SELECT *
			FROM binares_data
			WHERE
			  name = $14
			  AND uid = $15)`

var SyncUpdateCardTable = `WITH updated_rows AS (
				UPDATE cards
				SET
				  name = $1,
				  number = $2,
				  date = $3,
				  cvv = $4,
				  uId = $5,
				  deleted = $6,
				  last_update = $7
				WHERE
				  name = $8
				  AND last_update < $9
				  AND uid = $10
				RETURNING *
			  )
			  INSERT INTO cards (name, number, date, cvv, uid, deleted, last_update)
			  SELECT
				$11,
				$12,
				$13,
				$14,
				$15,
				$16,
				$17
			  WHERE
				NOT EXISTS (SELECT 1 FROM updated_rows)
				AND NOT EXISTS (SELECT *
				FROM cards
				WHERE
				  name = $18
				  AND uid = $19)`

var SynceTextTableActual = `SELECT name, data, uid, deleted, last_update
				  FROM text_data
				  WHERE
					name = $1
					AND last_update > $2`

var SynceLoginsTableActual = `SELECT name, login, password, uid, deleted, last_update
					FROM logins
					WHERE
					  name = $1
					  AND last_update > $2`

var SynceBinTableActual = `SELECT name, data, uid, deleted, last_update
					  FROM binares_data
					  WHERE
						name = $1
						AND last_update > $2`

var SynceCardTableActual = `SELECT name, number, date, cvv, uid, deleted, last_update
					  FROM cards
					  WHERE
						name = $1
						AND last_update > $2`

var SynceNewTextData = `SELECT name, data, uid, deleted, last_update FROM text_data WHERE uid = $1 AND name NOT IN (%s)`
var SynceNewBinData = `SELECT name, data, uid, deleted, last_update FROM binares_data WHERE uid = $1 AND name NOT IN (%s)`
var SynceNewAuthData = `SELECT name, login, password, uid, deleted, last_update FROM logins WHERE uid = $1 AND name NOT IN (%s)`
var SynceNewCardData = `SELECT name, number, date, cvv, uid, deleted, last_update FROM cards WHERE uid = $1 AND name NOT IN (%s)`
