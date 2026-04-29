package model

// User — учётная запись на сервере
type User struct {
	// ID — стабильный первичный ключ пользователя (непрозрачная строка, например UUID в виде текста).
	ID string
	// Login — уникальный человекочитаемый идентификатор (email или username — продуктовый выбор).
	Login string
	// PasswordHash — результат безопасного хэширования пользовательского пароля.
	PasswordHash []byte
}
