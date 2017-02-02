# Hack to get bootstrap working

window.jQuery = jQuery = require 'jquery'
Bootstrap  = require 'bootstrap'
React      = require "react"

Routes     = require './routes.cjsx'
Dispatcher = require "./dispatchers/Dispatcher"
