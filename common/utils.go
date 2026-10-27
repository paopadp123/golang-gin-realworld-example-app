// Common tools and helper functions
package common

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// A helper function to generate random string
func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		randIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		b[i] = letters[randIdx.Int64()]
	}
	return string(b)
}

// A helper function to generate random int
func RandInt() int {
	randNum, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		panic(err)
	}
	return int(randNum.Int64())
}

// Sentinel used by the user validator to keep the current password; it is not a secret
const RandomPassword = "A String Very Very Very Random!!@##$!@#4" // #nosec G101

// JWTSecretEnv is the environment variable that holds the HS256 signing secret
const JWTSecretEnv = "JWT_SECRET"

// minJWTSecretLen is the minimum secret length: 32 bytes (256 bits) for HS256
const minJWTSecretLen = 32

var jwtSecret []byte

// LoadJWTSecret reads the signing secret from JWT_SECRET.
// There is no default: the server must refuse to start without a real secret.
func LoadJWTSecret() error {
	secret := os.Getenv(JWTSecretEnv)
	if len(secret) < minJWTSecretLen {
		return fmt.Errorf("%s must be set to a random value of at least %d bytes", JWTSecretEnv, minJWTSecretLen)
	}
	jwtSecret = []byte(secret)
	return nil
}

// JWTSecret returns the secret loaded by LoadJWTSecret
func JWTSecret() []byte {
	return jwtSecret
}

// A Util function to generate jwt_token which can be used in the request header
func GenToken(id uint) string {
	jwt_token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	if len(JWTSecret()) == 0 {
		fmt.Printf("failed to sign JWT token for id %d: %s is not loaded\n", id, JWTSecretEnv)
		return ""
	}
	// Sign and get the complete encoded token as a string
	token, err := jwt_token.SignedString(JWTSecret())
	if err != nil {
		fmt.Printf("failed to sign JWT token for id %d: %v\n", id, err)
		return ""
	}
	return token
}

// My own Error type that will help return my customized Error info
//
//	{"database": {"hello":"no such table", error: "not_exists"}}
type CommonError struct {
	Errors map[string]interface{} `json:"errors"`
}

// To handle the error returned by c.Bind in gin framework
// https://github.com/go-playground/validator/blob/v9/_examples/translations/main.go
func NewValidatorError(err error) CommonError {
	res := CommonError{}
	res.Errors = make(map[string]interface{})
	var errs validator.ValidationErrors
	if !errors.As(err, &errs) {
		// Malformed JSON or a wrong type: not a validation error, and must not panic
		res.Errors["body"] = "invalid request body"
		return res
	}
	for _, v := range errs {
		// can translate each error one at a time.
		//fmt.Println("gg",v.NameNamespace)
		if v.Param() != "" {
			res.Errors[v.Field()] = fmt.Sprintf("{%v: %v}", v.Tag(), v.Param())
		} else {
			res.Errors[v.Field()] = fmt.Sprintf("{key: %v}", v.Tag())
		}

	}
	return res
}

// BindErrorStatus returns 422 for validation errors and 400 for anything else
// returned by Bind (malformed JSON, wrong types), which is a client error too
func BindErrorStatus(err error) int {
	var errs validator.ValidationErrors
	if errors.As(err, &errs) {
		return http.StatusUnprocessableEntity
	}
	return http.StatusBadRequest
}

// Wrap the error info in an object
func NewError(key string, err error) CommonError {
	res := CommonError{}
	res.Errors = make(map[string]interface{})
	res.Errors[key] = err.Error()
	return res
}

// Changed the c.MustBindWith() ->  c.ShouldBindWith().
// I don't want to auto return 400 when error happened.
// origin function is here: https://github.com/gin-gonic/gin/blob/master/context.go
func Bind(c *gin.Context, obj interface{}) error {
	b := binding.Default(c.Request.Method, c.ContentType())
	return c.ShouldBindWith(obj, b)
}
