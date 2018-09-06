'use strict';

const express = require('express');
const morgan = require('morgan');
const bodyParser = require('body-parser');
const httpStatus = require('http-status-codes');
const fabricQueryModule = require('api-modules/fabric/query');
const fs = require('fs');
const app = express();

// view engine setup
app.set('views', __dirname + '/views');
app.set('view engine', 'ejs');

// router
const query = require('./routes/query');
const invoke = require('./routes/invoke');

// css
const myCss = {
  style : fs.readFileSync('./api/views/index.css', 'utf8')
};

// body-parserを追加、JSONパース
app.use(bodyParser.urlencoded({extended: true}));
app.use(bodyParser.json());

app.use(morgan('combined'));

// routing
app.get('/', async function(req, res){
  const fabricModule = new fabricQueryModule("kawaya");
  const result = await fabricModule.run("getAllRooms", []);
  var rooms = result.rooms;
  res.render('index', {rooms: rooms, myCss: myCss});
});

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
