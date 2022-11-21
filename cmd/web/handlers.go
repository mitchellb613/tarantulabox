package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellb613/tarantulabox.git/internal/models"
	"github.com/mitchellb613/tarantulabox.git/internal/validator"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "home.html", data)
}

type userSignupForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, http.StatusOK, "signup.html", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	err = app.users.Insert(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, http.StatusOK, "login.html", data)
}
func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "You've been logged in successfully!")
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type tarantulaCreateForm struct {
	Species             string `form:"species"`
	Name                string `form:"name"`
	Next_Feed_Date      string `form:"next_feed_date"`
	Timezone_Offset     string `form:"timezone_offset"`
	Feed_Interval_Days  int    `form:"feed_interval_days"`
	Notify              bool   `form:"notify"`
	validator.Validator `form:"-"`
}

func (app *application) tarantulaCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = tarantulaCreateForm{
		Notify: false,
	}
	app.render(w, http.StatusOK, "create.html", data)
}

func (app *application) tarantulaCreatePost(w http.ResponseWriter, r *http.Request) {
	var form tarantulaCreateForm
	// r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE) //doesn't do anything?? using enforceMaxRequestSize middleware for now instead
	err := app.decodeMultipartPostForm(r, &form, MAX_UPLOAD_SIZE)
	if err != nil {
		app.serverError(w, err)
		return
	}

	form.CheckField(validator.NotBlank(form.Species), "species", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Next_Feed_Date), "next_feed_date", "This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Notify, true, false), "notify", "This field must true or false")
	form.CheckField(validator.IsPositive(form.Feed_Interval_Days), "feed_interval_days", "This field must be a positive value")
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.html", data)
		return
	}

	feed_date, err := time.Parse("2006-01-02T15:04Z07:00", form.Next_Feed_Date+form.Timezone_Offset)
	if err != nil {
		app.serverError(w, err)
		return
	}
	form.CheckField(validator.NotPast(feed_date), "next_feed_date", "This field cannot be a past date")
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.html", data)
		return
	}

	file, handler, err := r.FormFile("tarantula_image")
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer file.Close()

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		app.serverError(w, err)
		return
	}

	filetype := http.DetectContentType(buff)
	if filetype != "image/jpeg" && filetype != "image/png" {
		http.Error(w, "The provided file format is not allowed. Please upload a JPEG or PNG image", http.StatusBadRequest)
		return
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		app.serverError(w, err)
		return
	}

	owner_id := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)

	owner_id_string := strconv.Itoa(owner_id)

	err = os.MkdirAll("./uploads/"+owner_id_string, os.ModePerm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fileExtension := "." + strings.Split(handler.Filename, ".")[1]
	newFileName := randSeq(24) + fileExtension
	newFileURL := "/uploads/" + owner_id_string + "/" + newFileName

	dst, err := os.Create("." + newFileURL)
	defer dst.Close()
	if err != nil {
		app.serverError(w, err)
		return
	}

	if _, err := io.Copy(dst, file); err != nil {
		app.serverError(w, err)
		return
	}

	_, err = app.tarantulas.Insert(form.Species, form.Name, form.Feed_Interval_Days, form.Notify, newFileURL, owner_id, feed_date)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.resetTimer()
	app.sessionManager.Put(r.Context(), "flash", "Tarantula successfully created!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
