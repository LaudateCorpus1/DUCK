package ducklib

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"log"

	"github.com/Microsoft/DUCK/backend/ducklib/structs"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

var goPath = os.Getenv("GOPATH")
var testData = filepath.Join(goPath, "/src/github.com/Microsoft/DUCK/testdata.json")

//route Handlers

func helloHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Hello World")
}

func getDocSummaries(c echo.Context) error {

	docs, err := datab.GetDocumentSummariesForUser(c.Param("userid"))

	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, docs)
}

func testdataHandler(c echo.Context) error {

	dat, err := ioutil.ReadFile(testData)

	var e string
	if err != nil {
		e = err.Error()
		return c.JSON(http.StatusExpectationFailed, structs.Response{Ok: false, Reason: &e})

	}
	if err := FillTestdata(dat); err != nil {
		e = err.Error()
		return c.JSON(http.StatusConflict, structs.Response{Ok: false, Reason: &e})

	}

	return c.JSON(http.StatusOK, structs.Response{Ok: true})
}

/*
Document handlers
*/
func getDocHandler(c echo.Context) error {
	doc, err := datab.GetDocument(c.Param("docid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	fmt.Printf("GET revision: %s\n", doc.Revision)
	return c.JSON(http.StatusOK, doc)
}

func copyDocHandler(c echo.Context) error {
	doc, err := datab.GetDocument(c.Param("docid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	newDoc := new(structs.Document)
	if err := c.Bind(newDoc); err != nil {
		e := err.Error()
		fmt.Printf("Error at 1: %s\n", err)
		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}
	newDoc.Statements = doc.Statements

	id, err := datab.PostDocument(*newDoc)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	returnDoc, err := datab.GetDocument(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, returnDoc)
}

func deleteDocHandler(c echo.Context) error {
	err := datab.DeleteDocument(c.Param("docid"))
	if err != nil {
		e := err.Error()
		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}

	return c.JSON(http.StatusOK, structs.Response{Ok: true})
}

func putDocHandler(c echo.Context) error {
	/*
		resp, err := ioutil.ReadAll(c.Request().Body())
		if err != nil {
			e := err.Error()
			return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
		}
		fmt.Println(string(resp))
		return c.JSON(http.StatusNotFound, structs.Response{Ok: false})
	*/
	doc := new(structs.Document)
	if err := c.Bind(doc); err != nil {
		e := err.Error()
		fmt.Printf("Error at 1: %s\n", err)
		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}
	fmt.Printf("PUT revision: %s\n", doc.Revision)
	err := datab.PutDocument(*doc)
	if err != nil {
		e := err.Error()
		fmt.Printf("Error at 2: %s\n", err)
		if e == "Document update conflict." {
			return c.JSON(http.StatusConflict, structs.Response{Ok: false, Reason: &e})
		}
		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}

	docu, err := datab.GetDocument(doc.ID)
	fmt.Printf("PUT RETURN revision: %s\n", docu.Revision) // should be the same one we once got through the document GET
	if err != nil {
		e := err.Error()
		fmt.Printf("Error at 3: %s\n", err)

		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}

	return c.JSON(http.StatusOK, docu)
}
func postDocHandler(c echo.Context) error {

	doc := new(structs.Document)
	if err := c.Bind(doc); err != nil {
		e := err.Error()
		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}

	id, err := datab.PostDocument(*doc)
	if err != nil {
		e := err.Error()
		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}
	docu, err := datab.GetDocument(id)
	if err != nil {
		e := err.Error()
		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}

	return c.JSON(http.StatusOK, docu)
}

/*
User handlers
*/

func deleteUserHandler(c echo.Context) error {
	err := datab.DeleteUser(c.Param("id"))
	if err != nil {
		e := err.Error()
		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}

	return c.JSON(http.StatusOK, structs.Response{Ok: true})
}

func putUserHandler(c echo.Context) error {

	u := new(structs.User)
	if err := c.Bind(u); err != nil {
		e := err.Error()
		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}
	id := c.Param("id")
	u.ID = id
	err := datab.PutUser(*u)
	if err != nil {
		e := err.Error()
		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}

	us, err := datab.GetUser(id)
	if err != nil {
		e := err.Error()
		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}

	return c.JSON(http.StatusOK, us)
}
func postUserHandler(c echo.Context) error {
	newUser := new(structs.User)
	if err := c.Bind(newUser); err != nil {
		e := err.Error()
		log.Println(e)
		return c.JSON(http.StatusBadRequest, structs.Response{Ok: false, Reason: &e})
	}

	id, err := datab.PostUser(*newUser)
	if err != nil {
		e := err.Error()
		log.Println(e)
		return c.JSON(http.StatusInternalServerError, structs.Response{Ok: false, Reason: &e})
	}
	var u, err2 = datab.GetUser(id)
	if err2 != nil {
		e := err2.Error()
		log.Println(e)
		return c.JSON(http.StatusInternalServerError, structs.Response{Ok: false, Reason: &e})
	}

	return c.JSON(http.StatusOK, u)
}

/*
Rulebase handlers
*/

//DB Rulebasehandlers are not used
/*
func deleteRsHandler(c echo.Context) error {
	err := datab.DeleteRulebase(c.Param("id"))
	if err != nil {
		e := err.Error()
		return c.JSON(http.StatusNotFound, Response{Ok: false, Reason: &e})
	}

	return c.JSON(http.StatusOK, Response{Ok: true})
}

func putRsHandler(c echo.Context) error {

	resp, err := ioutil.ReadAll(c.Request().Body())
	if err != nil {
		e := err.Error()
		return c.JSON(http.StatusNotFound, Response{Ok: false, Reason: &e})
	}
	id := c.Param("id")
	err = datab.PutRulebase(id, resp)
	if err != nil {
		e := err.Error()
		return c.JSON(http.StatusNotFound, Response{Ok: false, Reason: &e})
	}

	doc, err := datab.GetRulebase(id)
	if err != nil {
		e := err.Error()
		return c.JSON(http.StatusNotFound, Response{Ok: false, Reason: &e})
	}

	return c.JSON(http.StatusOK, doc)
}

func postRsHandler(c echo.Context) error {

	req, err := ioutil.ReadAll(c.Request().Body())
	if err != nil {
		e := err.Error()
		return c.JSON(http.StatusNotFound, Response{Ok: false, Reason: &e})
	}

	id, err := datab.PostRulebase(req)
	if err != nil {
		e := err.Error()
		return c.JSON(http.StatusNotFound, Response{Ok: false, Reason: &e})
	}
	doc, err := datab.GetRulebase(id)
	if err != nil {
		e := err.Error()
		return c.JSON(http.StatusNotFound, Response{Ok: false, Reason: &e})
	}

	return c.JSON(http.StatusOK, doc)
}
*/
func checkDocHandler(c echo.Context) error {
	/*
		resp, err := ioutil.ReadAll(c.Request().Body())
		if err != nil {
			e := err.Error()
			return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
		}
		fmt.Println(string(resp))
		return c.JSON(http.StatusNotFound, structs.Response{Ok: false})
	*/
	id := c.Param("baseid")

	doc := new(structs.Document)
	if err := c.Bind(doc); err != nil {
		e := err.Error()

		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}

	ok, docs, err := checker.CompliantDocuments(id, doc, 10, 0)
	if err != nil {
		e := err.Error()

		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}

	if ok {

		return c.JSON(http.StatusOK, structs.ComplianceResponse{Ok: ok, Compliant: "COMPLIANT"})
	}

	return c.JSON(http.StatusOK, structs.ComplianceResponse{Ok: ok, Compliant: "NON_COMPLIANT", Documents: docs})
}

func checkDocIDHandler(c echo.Context) error {
	id := c.Param("baseid")
	docid := c.Param("documentid")

	doc, err := datab.GetDocument(docid)
	if err != nil {
		e := err.Error()

		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}

	ok, docs, err := checker.CompliantDocuments(id, &doc, 10, 0)
	if err != nil {
		e := err.Error()

		return c.JSON(http.StatusNotFound, structs.Response{Ok: false, Reason: &e})
	}
	fmt.Printf("Compliant: %t", ok)
	if ok {
		return c.JSON(http.StatusOK, structs.ComplianceResponse{Ok: ok, Compliant: "COMPLIANT"})
	}

	return c.JSON(http.StatusOK, structs.ComplianceResponse{Ok: ok, Compliant: "NON_COMPLIANT", Documents: docs})

}

// loginHandler handles the login Process
func loginHandler(c echo.Context) error {

	u := new(structs.Login)
	if err := c.Bind(u); err != nil {
		return err
	}
	id, pw, err := datab.GetLogin(u.Username) //TODO compare with encrypted pw
	if err != nil {
		log.Println(err)
		e := err.Error()

		return c.JSON(http.StatusUnauthorized, structs.Response{Ok: false, Reason: &e})
	}
	log.Printf("id: %s, pw: %s", id, pw)
	if u.Password == pw {

		user, err := datab.GetUser(id)
		if err != nil {
			log.Println(err)
			return echo.ErrUnauthorized
		}

		// Create token
		token := jwt.New(jwt.SigningMethodHS256)

		// Set claims
		token.Claims["firstName"] = user.Firstname
		token.Claims["lastName"] = user.Lastname
		token.Claims["id"] = user.ID
		token.Claims["permissions"] = 1024 //FIXME
		token.Claims["exp"] = time.Now().Add(time.Hour * 24 * 7).Unix()

		// Generate encoded token and send it as response.
		t, err := token.SignedString([]byte(JWT))
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]string{
			"token":     t,
			"firstName": user.Firstname,
			"lastName":  user.Lastname,
			"id":        user.ID,
			"locale":    user.Locale,
		})
	}
	log.Println("Passwords do not match")
	return echo.ErrUnauthorized

}
