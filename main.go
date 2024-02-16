package main

import (
	"fmt"
	"html/template"
	"net/http" //Функциональность для работы с HTTP-запросами
)

type Rsvp struct { //создание пользовательского типа данных
	Name, Email, Phone string
	WillAttend         bool
}

var responses = make([]*Rsvp, 0, 10) //срез указателей на экземпляры структуры Rsvp

/*
	карта использует string ключи,

используемые для хранения указателей на экземпляры структуры Template, определенной
пакетом html/template.
*/
var templates = make(map[string]*template.Template, 3)

func loadTemplates() { //отвечает за загрузку файлов HTML
	templateNames := [5]string{"welcome", "form", "thanks", "sorry", "list"} //массив имен файлов
	for index, name := range templateNames {                                 //перебор массива имен

		/* Пакет html/templates предоставляет функцию ParseFiles, которая используется для
		загрузки и обработки HTML-файлов. Одной из самых полезных и необычных возможностей Go
		является то, что функции могут возвращать несколько результирующих значений. Функция
		ParseFiles возвращает два результата: указатель на значение template.Template и ошибку,
		которая является встроенным типом данных для представления ошибок в Go. Краткий синтаксис
		для создания переменных используется для присвоения этих двух результатов переменным */
		t, err := template.ParseFiles("layout.html", name+".html")

		/* Если err равен nil, я добавляю на карту пару ключ-значение, используя значение name в
		качестве ключа и *template.Tempate, назначенный t в качестве значения. */
		if err == nil {
			templates[name] = t
			fmt.Println("Loaded template", index, name)
		} else {
			panic(err)
		}
	}
}

/*
	Функциональность для работы с HTTP-запросами

Второй аргумент — это указатель на экземпляр структуры Request, определенной в пакете
net/http, который описывает обрабатываемый запрос. Первый аргумент — это пример
интерфейса, поэтому он не определен как указатель
*/
func welcomeHandler(writer http.ResponseWriter, request *http.Request) {
	templates["welcome"].Execute(writer, nil)
}

/*
	Этот оператор считывает *template.Template из карты, назначенной переменной templates,

и вызывает определенный им метод Execute. Первый аргумент — это ResponseWriter, куда будут
записываться выходные данные ответа, а второй аргумент — это значение данных, которое можно
использовать в выражениях, содержащихся в шаблоне.
*/
func listHandler(writer http.ResponseWriter, request *http.Request) {
	templates["list"].Execute(writer, responses)
}

/*
	В результате структуру formData можно использовать так, как будто она определяет поля Name,

Email, Phone и WillAttend из структуры Rsvp, и я могу создать экземпляр структуры formData,
используя существующее значение Rsvp. Звездочка обозначает указатель, что означает, что я не
хочу копировать значение Rsvp при создании значения formData.
*/
type formData struct {
	*Rsvp
	Errors []string
}

func formHandler(writer http.ResponseWriter, request *http.Request) {
	/* проверяет значение поля request.Method, которое возвращает тип
	полученного HTTP-запроса. Для GET-запросов выполняется шаблон form */
	if request.Method == http.MethodGet {
		templates["form"].Execute(writer, formData{
			/* Нет данных для использования при ответе на запросы GET, но мне нужно предоставить
			шаблон с ожидаемой структурой данных. Для этого я создаю экземпляр структуры formData,
			используя значения по умолчанию для ее полей */
			Rsvp: &Rsvp{}, Errors: []string{},
		})
	} else if request.Method == http.MethodPost {
		/* Метод ParseForm обрабатывает данные формы, содержащиеся в HTTP-запросе, и заполняет
		карту, доступ к которой можно получить через поле Form. Затем данные формы используются для
		создания значения Rsvp */
		request.ParseForm()
		responseData := Rsvp{
			Name:       request.Form["name"][0],
			Email:      request.Form["email"][0],
			Phone:      request.Form["phone"][0],
			WillAttend: request.Form["willattend"][0] == "true",
		}
		errors := []string{}
		if responseData.Name == "" {
			errors = append(errors, "Please enter your name")
		}
		if responseData.Email == "" {
			errors = append(errors, "Please enter your email address")
		}
		if responseData.Phone == "" {
			errors = append(errors, "Please enter your phone number")
		}
		if len(errors) > 0 {
			templates["form"].Execute(writer, formData{
				Rsvp: &responseData, Errors: errors,
			})
		} else {
			// Создав значение Rsvp, я добавляю его в срез, присвоенный переменной responses
			responses = append(responses, &responseData)

			if responseData.WillAttend {
				templates["thanks"].Execute(writer, responseData.Name)
			} else {
				templates["sorry"].Execute(writer, responseData.Name)
			}
		}
	}
}
func main() {
	loadTemplates()

	/* используется для указания URLадреса и обработчика, который будет получать соответствующие запросы. Я использовал
	HandleFunc для регистрации своих новых функций-обработчиков, чтобы они реагировали на
	URL-пути / и /list: */
	http.HandleFunc("/", welcomeHandler)
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/form", formHandler)

	/* операторы создают HTTP-сервер, который прослушивает запросы через порт 5000,
	указанный первым аргументом функции ListenAndServe. Второй аргумент равен nil, что говорит
	серверу, что запросы должны обрабатываться с использованием функций, зарегистрированных с
	помощью функции HandleFunc */
	err := http.ListenAndServe(":5000", nil)
	if err != nil {
		fmt.Println(err)
	}
}
