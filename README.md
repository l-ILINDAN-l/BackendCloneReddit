# A pet project to write a backend for the Reddit clan with a ready-made frontend

The API returns responses with JSON files

## The following API methods are implemented:

1) POST /api/register - registration
2) POST /api/login - login
3) GET /api/posts/ - list of all posts
4) POST /api/posts/ - adding a post - please note - there is an url, but there is a text
5) GET /api/posts/{CATEGORY_NAME} - a list of posts of a specific category
6) GET /api/post/{POST_ID} - details of the post with comments
7) POST /api/post/{POST_ID} - adding a comment
8) DELETE /api/post/{POST_ID}/{COMMENT_ID} - delete a comment
9) GET /api/post/{POST_ID}/upvote - rating the post up
10) GET /api/post/{POST_ID}/downvote - the rating of the post is down
11) GET /api/post/{POST_ID}/unvote - voice cancellation 
12) DELETE /api/post/{POST_ID} - deleting a post
13) GET /api/user/{USER_LOGIN} - getting all posts of a specific user

## Inside you will have the following models:

1) Comments on posts
2) The Post
3) Session
4) The user
5) Vote for the post

## There are also interfaces for working with databases that store model objects.
1) UserRepository
2) SessionRepository
3) PostRepository

The project provides a simplification in view of the fact that data is stored in memory.
