### Whole app
-- Improve error response from the backend to the frontend

### Middleware
1 -- should rate limit requests to prevent over load and brute forcing

2 -- Verrify human and not a bot through and external tool


## additional routes to add
a change user name route
a change email route




## modifications for existing routes
# Sign up a user

1 -- sign up user should have a step in where the user must verify their email (backend)

1 -- token should be stored as a hash not plain text

2 -- token should be displayed once after signing up for use

3 -- sign up user should send and email for admin approval or require payment in the future (backend)



# Cycle User API Token (for future use)

1 -- token should be stored as a hash not plain text

1 -- move the form from a single page in the front end into the user profile page on the front end

#  Forgot password final reset

1 -- have the front end success message -> instead redirect to the dashboard of the user





# List detail

1 -- pagify messages and subscribers for returned

# List index

1 -- filter the output from the index so less data is returned