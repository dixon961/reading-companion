package service

import (
	"strings"
	"testing"
	
	"github.com/stretchr/testify/assert"
)

// mockFile is a simple implementation of multipart.File for testing
type mockFile struct {
	*strings.Reader
}

func (m *mockFile) Close() error {
	return nil
}

// TestSessionService_validateHighlights tests the validateHighlights method
func TestSessionService_validateHighlights(t *testing.T) {
	service := &SessionService{}
	
	// Test with valid highlights (more than 3)
	validHighlights := []string{"Highlight 1", "Highlight 2", "Highlight 3", "Highlight 4"}
	err := service.validateHighlights(validHighlights)
	assert.NoError(t, err)
	
	// Test with invalid highlights (less than 3)
	invalidHighlights := []string{"Highlight 1", "Highlight 2"}
	err = service.validateHighlights(invalidHighlights)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "minimum 3 highlights required")
	
	// Test with empty highlights
	emptyHighlights := []string{}
	err = service.validateHighlights(emptyHighlights)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "minimum 3 highlights required")
}

// TestSessionService_parseHighlights tests the parseHighlights method
func TestSessionService_parseHighlights(t *testing.T) {
	service := &SessionService{}
	
	// Test with valid highlights
	validContent := "First highlight\n\nSecond highlight\n\nThird highlight"
	reader := &mockFile{strings.NewReader(validContent)}
	highlights, err := service.parseHighlights(reader)
	
	assert.NoError(t, err)
	assert.Len(t, highlights, 3)
	assert.Equal(t, "First highlight", highlights[0])
	assert.Equal(t, "Second highlight", highlights[1])
	assert.Equal(t, "Third highlight", highlights[2])
	
	// Test with empty content
	emptyReader := &mockFile{strings.NewReader("")}
	highlights, err = service.parseHighlights(emptyReader)
	
	assert.NoError(t, err)
	assert.Len(t, highlights, 0)
	
	// Test with content that has extra newlines
	extraNewlines := "\n\nFirst highlight\n\n\n\nSecond highlight\n\n\n"
	reader = &mockFile{strings.NewReader(extraNewlines)}
	highlights, err = service.parseHighlights(reader)
	
	assert.NoError(t, err)
	assert.Len(t, highlights, 2)
	assert.Equal(t, "First highlight", highlights[0])
	assert.Equal(t, "Second highlight", highlights[1])
	
	// Test with kon-tiki format
	konTikiContent := `Заметка BOOX | <<Высоконагруженные приложения. Программирование, масштабирование, поддержка - Мартин Клеппман>>
Kon_Tiki2

время：2025-07-07 20:03
【Контент】По сути, нам нужна предикатная блокировка (predicate 
lock) [3]. Она аналогична описанным ранее разделяемым/монопольным блокировкам, но относится не к конкретному объекту (например, одной строке в таблице), 
а ко всем объектам, удовлетворяющим какому-то условию отбора:
SELECT * FROM bookings
 WHERE room_id = 123 AND
 end_time > '2018-01-01 12:00' AND
 start_time < '2018-01-01 13:00';
Предикатная блокировка ограничивает доступ следующим образом.
【Заметки】1
-------------------

время：2025-07-07 20:04
【Контент】идея заключается в применимости предикатных блокировок даже к тем 
объектам, которые еще не существуют в базе данных, но могут там появиться в будущем (фантомы). Если двухфазная блокировка включает предикатные блокировки, 
то база данных предотвращает все формы асимметрии записи и других состояний 
гонки, так что ее изоляцию можно с полным правом назвать сериализуемостью.
【Заметки】2
-------------------

время：2025-07-07 20:06
【Контент】Можно спокойно упростить предикат, чтобы он соответствовал более широкому 
множеству объектов. Например, расширить предикатную блокировку для бронирования конференц-зала номер 123 с полудня до часа дня, блокировав бронирование
【Заметки】3
-------------------

время：2025-07-07 20:06
【Контент】зала 123 на любое время, или все залы (а не только 123) с полудня до часа дня. 
Это вполне безопасно, ведь любые подходящие под изначальный предикат операции записи определенно будут соответствовать также и расширенной версии.
В базе бронирования конференц-залов будет, вероятно, индекс по столбцу room_id
и/или индексы по столбцам start_time и end_time (в противном случае вышеприведенный запрос станет работать на большой базе данных очень медленно
【Заметки】
-------------------

время：2025-07-07 20:08
【Контент】Пессимистическое и оптимистическое 
управление конкурентным доступом
Двухфазная блокировка представляет собой так называемый механизм пессимистического управления конкурентным доступом: он основан на следующем 
принципе: если что-то может потенциально пойти не так (о чем сигнализирует 
удерживаемая другой транзакцией блокировка), то лучше подождать нормализации ситуации, прежде чем выполнять какие-либо действия. Это напоминает взаимоисключающие блокировки (mutual exclusion), используемые для защиты данных 
в многопоточном программировании.
【Заметки】5
-------------------

время：2025-07-07 20:09
【Контент】Напротив, сериализуемая изоляция снимков состояния представляет собой оптимистический метод управления конкурентным доступом. «Оптимистический» 
в этом контексте означает следующее: вместо блокировки в случае потенциально 
опасных действий транзакции просто продолжают выполняться в надежде, что все 
будет хорошо. При фиксации транзакции база данных проверяет, не случилось ли 
чего-то плохого (например, не была ли нарушена изоляция). Если да, то транзакция 
прерывается и ее выполнение приходится повторять еще раз. Допускается фиксация только выполненных сериализуемым образом транзакций
【Заметки】6
-------------------`
	
	konTikiReader := &mockFile{strings.NewReader(konTikiContent)}
	highlights, err = service.parseHighlights(konTikiReader)
	
	assert.NoError(t, err)
	assert.Len(t, highlights, 6) // Should have 6 highlights from the kon-tiki format
	
	// Check that we have the expected number of highlights
	assert.Len(t, highlights, 6)
	
	// Check the content of the first highlight starts with expected text
	assert.Contains(t, highlights[0], "По сути, нам нужна предикатная блокировка (predicate lock) [3]")
	
	// Check the content of the second highlight
	assert.Contains(t, highlights[1], "идея заключается в применимости предикатных блокировок даже к тем объектам")
}