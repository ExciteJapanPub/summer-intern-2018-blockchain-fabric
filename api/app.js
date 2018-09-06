'use strict';

const express = require('express');
const morgan = require('morgan');
const bodyParser = require('body-parser');
const httpStatus = require('http-status-codes');
const fabricQueryModule = require('api-modules/fabric/query');
const fabricInvokeModule = require('api-modules/fabric/invoke');
const app = express();

// view engine setup
app.set('views', __dirname + '/views');
app.set('view engine', 'ejs');
app.use('/static', express.static('./api/views'));

// router
const query = require('./routes/query');
const invoke = require('./routes/invoke');

// body-parserを追加、JSONパース
app.use(bodyParser.urlencoded({extended: true}));
app.use(bodyParser.json());

app.use(morgan('combined'));

// routing
app.get('/', async function(req, res){
  const fabricModule = new fabricQueryModule("kawaya");
  const result = await fabricModule.run("getAllRooms", []);
  var rooms = result.rooms;
  res.render('index', {rooms: rooms});
});

// routing
app.post('/', async function(req, res){
  let result;
  const invokeModule = new fabricInvokeModule("kawaya");
  const params = [req.body.user_hash, req.body.room_id];
  result = await invokeModule.run("reserve", params);
  if(result.status !== 200){
    res.render("error");
    return;
  }
  const queryModule = new fabricQueryModule("kawaya");
  result = await queryModule.run("getAllRooms", []);
  var rooms = result.rooms;
  res.render('index', {rooms: rooms});
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
