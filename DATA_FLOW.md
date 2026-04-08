### Data Flow
The following file will outline how the data flows through out the appliaction and how things chain together

## Different types of structures:
Users (account holders)
Lists (The core structure of the application)
Contacts (Third party members who either recieve or send mail to Users)
Messages (Structure to hold the mail sent)

# Sign up a user
[User]
v
[Frontend]
domain-front/signup
- user enters email and password
- user can also enter a username (optional)
v
[Backend]
domain-back/api/user/register
- validate email
- make sure it is unique
- validate password strength (special char, number, upper, lower, at least 8 chars long)
- store information + generate api token (for later use)
- assign the jwt as a cookie to track the session (expires after 24 hrs)
v
[response]
"api_token"
"is_admin"
"user_email"
"user_name"
(jwt as cookie)

# Login as user
[User]
v
[Frontend]
domain-front/api/user/login
- user enters email and password
v
[Backend]
domain-back/login
- validate email
- make sure it is in the database
- validate password input matches the password stored in the database
- assign a new jwt as a cookie to track the session (expires after 24 hrs)
v
[response]
"is_admin"
"user_email"
"user_name"
(jwt as cookie)

# Cycle User API Token

[User] - must be autenticated already
v
[Frontend]
(inside of user profile page)
- user enters password + userID is retrieved from token
v
[Backend]
domain-back/api/user/cycle-token
- validate there is a user with UID passed in from the auth cookie
- validate password input matches the password stored in the database
- replace the old token with a new token therefore depricating the old
v
[response]
"api_token"

# User forgot password email request

[User]
v
[Frontend]
domain-front/forgot-password
- user enters email
v
[Backend]
domain-back/api/user/forgot-password
- validate email
- make sure it is in the database
- generate a token for use in the email containing the reset link
- send an email link to a front end form for next steps in password reset
v
[response]
"api_token"
v
[Email]
domain-front/reset-password/:token-for-reset


# User forgot password email request

[User] - must be autenticated already
v
[Email]
domain-front/reset-password/:token-for-reset
v
[Frontend]
domain/reset-password/:token-for-reset
- user enters new password (also sends token)
v
[Backend]
/api/user/change-password/:token
- validate there is a user with UID passed in from the auth cookie
- validate password input is valid and meets the criteria
- replace the old password with the new password
v
[response]
"message"

# User logout
[User]
v
[Frontend]
navbar of application
- user enters
v
[Backend]
domain-back/api/user/logout
- invalidate the cookie that was returned with the reponse
v
[response]
"message"

# GetMe 
[middleware]
checks to see if the user is a valid user or not

# Create a list
[User]
v
[Frontend]
domain-front/mailing-list/create
- user enters name, list_type, and a public_facing_name
v
[Backend]
domain-back/api/list/create
- validate unique and create
v
[response]
"name"
"list-type"
"user-ID"
"public_ID"
"public_facing_name"

# Delete list
[User]
v
[Frontend]
list detail page
- list id as a param
v
[Backend]
domain-back/api/list/delete/:id
- delete list 
v
[response]
"reponse"

# list detail
[User]
v
[Frontend]
list on list index
- takes list id as a param
v
[Backend]
domain-back/api/list/detail/:id
- fetch list data
v
[response]
------- list data

# list index
[User]
v
[Frontend]
list index page
- count_per_page and page passed in via a form
v
[Backend]
domain-back/api/list/index
- grab lists short handed view (using page count and page (number) as filter)
[response]
------- list of pages