package dbhandler

// Question Type
const (
	QuestionTypeTruth = iota
	QuestionTypeDare
	QuestionTypeTruth18
	QuestionTypeDare18
)

type Question struct {
	Id      int
	Type    int
	Content string
	GroupId int
}

func InsertQuestion(questionType int, content string, groupId int) (*Question, error) {
	if db == nil {
		return nil, ErrDBNotInitialized
	}

	stmt, err := db.Prepare("INSERT INTO Questions(Type, Content, GroupId) VALUES(?,?,?)")
	if err != nil {
		return nil, err
	}

	res, err := stmt.Exec(questionType, content, groupId)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	return &Question{int(id), questionType, content, groupId}, err
}

func ExistsQuestion(content string) (bool, error) {
	if db == nil {
		return false, ErrDBNotInitialized
	}

	rows, err := db.Query(`SELECT * FROM Questions WHERE Content="` + content + `"`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func UpdateQuestion(question *Question) error {
	if db == nil {
		return ErrDBNotInitialized
	}

	stmt, err := db.Prepare("UPDATE Questions SET Type=?, Content=?, GroupId=? WHERE Id=?")
	if err != nil {
		return err
	}

	res, err := stmt.Exec(question.Type, question.Content, question.GroupId, question.Id)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	return err
}

func RandomQuestion() (*Question, error) {
	rows, err := db.Query("SELECT * FROM Questions ORDER BY RANDOM() LIMIT 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	question := &Question{}
	rows.Next()
	err = rows.Scan(&question.Id, &question.Type, &question.Content, &question.GroupId)
	if err != nil {
		return nil, err
	}

	return question, nil
}
