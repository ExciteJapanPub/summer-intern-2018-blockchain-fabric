'use strict';

const express = require('express');
const morgan = require('morgan');
const bodyParser = require('body-parser');
const httpStatus = require('http-status-codes');
const app = express();


// router
const query = require('./routes/query');
const invoke = require('./routes/invoke');

// body-parserを追加、JSONパース
app.use(bodyParser.urlencoded({extended: true}));
app.use(bodyParser.json());

app.use(morgan('combined'));

// routing
app.use('/query', query);
app.use('/invoke', invoke);

// error handler
app.use(function(err, req, res, next) {
    if (err.msg !== undefined) {
        const status = err.status || httpStatus.INTERNAL_SERVER_ERROR;
        delete (err.status);

        res.status(status);
        res.jsonp(err);
        return;
    }
    next(err);
});

app.listen(3000);
