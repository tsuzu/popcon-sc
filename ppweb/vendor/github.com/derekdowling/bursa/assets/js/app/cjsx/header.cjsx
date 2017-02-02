# @cjsx React.DOM
React = require 'react'

module.exports = Header = React.createClass
  render: ->
    <section id="header" className="header-container">
      <header className="top-header clearfix">
          <div className="logo">
              <a href="#">
                  <span>Bursa</span>
              </a>
          </div>

          <div className="menu-button" toggle-off-canvas>
              <span className="icon-bar"></span>
              <span className="icon-bar"></span>
              <span className="icon-bar"></span>
          </div>

          <div className="top-nav">
              <ul className="nav-left list-unstyled">
                  <li>
                      <a href="#/" data-toggle-nav-collapsed-min
                                   className="toggle-min"
                                   id="step4"
                                   ><i className="fa fa-bars"></i></a>
                  </li>
                  <li className="dropdown hidden-xs">
                      <a href="javascript:;" className="dropdown-toggle" id="step1" data-toggle="dropdown"><i className="fa fa-cogs"></i></a>
                      <div className="dropdown-menu with-arrow panel panel-default admin-options" ui-not-close-on-click>
                          <div className="panel-heading"> Admin Options </div>
                          <ul className="list-group">
                              <li className="list-group-item">
                                  <p>Layouts Style</p>
                                  <label className="ui-radio"><input name="layout" type="radio" value="boxed" ng-model="admin.layout"><span>Boxed</span></label>
                                  <label className="ui-radio"><input name="layout" type="radio" value="wide" ng-model="admin.layout"><span>Wide</span></label>
                              </li>
                              <li className="list-group-item">
                                  <p>Menu Style</p>
                                  <label className="ui-radio"><input name="menu" type="radio" value="vertical" ng-model="admin.menu"><span>Vertical</span></label>
                                  <label className="ui-radio"><input name="menu" type="radio" value="horizontal" ng-model="admin.menu"><span>Horizontal</span></label>
                              </li>
                              <li className="list-group-item">
                                  <p>Additional</p>
                                  <label className="ui-checkbox"><input name="checkbox1" type="checkbox" value="option1" ng-model="admin.fixedHeader"><span>Fixed Top Header</span></label>
                                  <br>
                                  <label className="ui-checkbox"><input name="checkbox1" type="checkbox" value="option1" ng-model="admin.fixedSidebar"><span>Fixed Sidebar Menu</span></label>
                              </li>
                          </ul>
                      </div>
                  </li>
                  <li className="dropdown text-normal nav-profile">
                      <a href="javascript:;" className="dropdown-toggle" data-toggle="dropdown">
                          <span className="hidden-xs">
                              <span data-i18n="Lisa Doe"></span>
                          </span>
                          <img src="/img/g1.jpg" className="img-circle img30_30" />
                      </a>
                      <ul className="dropdown-menu with-arrow">
                          <li>
                              <a href="#/pages/profile">
                                  <i className="fa fa-user"></i>
                                  <span data-i18n="My Profile"></span>
                              </a>
                          </li>
                          <li>
                              <a href="#/tasks/tasks">
                                  <i className="fa fa-check"></i>
                                  <span data-i18n="My Tasks"></span>
                              </a>
                          </li>
                          <li>
                              <a href="#/pages/lock-screen">
                                  <i className="fa fa-lock"></i>
                                  <span data-i18n="Lock"></span>
                              </a>
                          </li>
                          <li>
                              <a href="#/pages/signin">
                                  <i className="fa fa-sign-out"></i>
                                  <span data-i18n="Log Out"></span>
                              </a>
                          </li>
                      </ul>
                  </li>
                  <li className="dropdown langs text-normal" data-ng-controller="LangCtrl">
                      <a href="javascript:;" className="dropdown-toggle active-flag" id="step2" data-toggle="dropdown">
                          <div className="flag">EN</div>
                      </a>
                      <ul className="dropdown-menu with-arrow list-langs" role="menu">
                          <li>
                              <a href="javascript:;" data-ng-click="setLang('English')"><div className="flag flags-american">EN</div> English</a></li>
                          <li data-ng-show="lang !== 'Español' ">
                              <a href="javascript:;" data-ng-click="setLang('Español')"><div className="flag flags-spain">ES</div> Español</a></li>
                      </ul>
                  </li>
                  <li className="dropdown">
                      <a href="javascript:;" className="dropdown-toggle" data-toggle="dropdown">
                          <span className="fa fa-magic nav-icon"></span>
                      </a>
                      <ul className="dropdown-menu pull-right color-switch" data-ui-color-switch>
                          <li>
                            <a href="javascript:;" className="color-option color-some_color" data-style="some_color"></a>
                          </li>
                      </ul>
                  </li>
              </ul>

              <ul className="nav-right pull-right list-unstyled">
                  <li className="dropdown">
                      <a href="javascript:;" className="dropdown-toggle bg-info" data-toggle="dropdown">
                          <i className="fa fa-comment-o"></i>
                          <span className="badge badge-info">2</span>
                      </a>
                      <div className="dropdown-menu pull-right with-arrow panel panel-default">
                          <div className="panel-heading">
                              You have 2 messages.
                          </div>
                          <ul className="list-group">
                              <li className="list-group-item">
                                  <a href="javascript:;" className="media">
                                      <span className="pull-left media-icon">
                                          <span className="round-icon sm bg-info"><i className="fa fa-comment-o"></i></span>
                                      </span>
                                      <div className="media-body">
                                          <span className="block">Jane sent you a message</span>
                                          <span className="text-muted">3 hours ago</span>
                                      </div>
                                  </a>
                              </li>
                              <li className="list-group-item">
                                  <a href="javascript:;" className="media">
                                      <span className="pull-left media-icon">
                                          <span className="round-icon sm bg-danger"><i className="fa fa-comment-o"></i></span>
                                      </span>
                                      <div className="media-body">
                                          <span className="block">Lynda sent you a mail</span>
                                          <span className="text-muted">9 hours ago</span>
                                      </div>
                                  </a>
                              </li>
                          </ul>
                          <div className="panel-footer">
                              <a href="javascript:;">Show all messages.</a>
                          </div>
                      </div>
                  </li>
                  <li className="dropdown">
                      <a href="javascript:;" className="dropdown-toggle bg-success" data-toggle="dropdown">
                          <i className="fa fa-envelope-o"></i>
                          <span className="badge badge-info">3</span>
                      </a>
                      <div className="dropdown-menu pull-right with-arrow panel panel-default">
                          <div className="panel-heading">
                              You have 3 mails.
                          </div>
                          <ul className="list-group">
                              <li className="list-group-item">
                                  <a href="javascript:;" className="media">
                                      <span className="pull-left media-icon">
                                          <span className="round-icon sm bg-warning"><i className="fa fa-envelope-o"></i></span>
                                      </span>
                                      <div className="media-body">
                                          <span className="block">Lisa sent you a mail</span>
                                          <span className="text-muted block">2min ago</span>
                                      </div>
                                  </a>
                              </li>
                              <li className="list-group-item">
                                  <a href="javascript:;" className="media">
                                      <span className="pull-left media-icon">
                                          <span className="round-icon sm bg-info"><i className="fa fa-envelope-o"></i></span>
                                      </span>
                                      <div className="media-body">
                                          <span className="block">Jane sent you a mail</span>
                                          <span className="text-muted">3 hours ago</span>
                                      </div>
                                  </a>
                              </li>
                              <li className="list-group-item">
                                  <a href="javascript:;" className="media">
                                      <span className="pull-left media-icon">
                                          <span className="round-icon sm bg-success"><i className="fa fa-envelope-o"></i></span>
                                      </span>
                                      <div className="media-body">
                                          <span className="block">Lynda sent you a mail</span>
                                          <span className="text-muted">9 hours ago</span>
                                      </div>
                                  </a>
                              </li>
                          </ul>
                          <div className="panel-footer">
                              <a href="javascript:;">Show all mails.</a>
                          </div>
                      </div>
                  </li>
                  <li className="dropdown">
                      <a href="javascript:;" className="dropdown-toggle bg-warning" data-toggle="dropdown">
                          <i className="fa fa-bell-o nav-icon"></i>
                          <span className="badge badge-info">3</span>
                      </a>
                      <div className="dropdown-menu pull-right with-arrow panel panel-default">
                          <div className="panel-heading">
                              You have 3 notifications.
                          </div>
                          <ul className="list-group">
                              <li className="list-group-item">
                                  <a href="javascript:;" className="media">
                                      <span className="pull-left media-icon">
                                          <span className="round-icon sm bg-success"><i className="fa fa-bell-o"></i></span>
                                      </span>
                                      <div className="media-body">
                                          <span className="block">New tasks needs to be done</span>
                                          <span className="text-muted block">2min ago</span>
                                      </div>
                                  </a>
                              </li>
                              <li className="list-group-item">
                                  <a href="javascript:;" className="media">
                                      <span className="pull-left media-icon">
                                          <span className="round-icon sm bg-info"><i className="fa fa-bell-o"></i></span>
                                      </span>
                                      <div className="media-body">
                                          <span className="block">Change your password</span>
                                          <span className="text-muted">3 hours ago</span>
                                      </div>
                                  </a>
                              </li>
                              <li className="list-group-item">
                                  <a href="javascript:;" className="media">
                                      <span className="pull-left media-icon">
                                          <span className="round-icon sm bg-danger"><i className="fa fa-bell-o"></i></span>
                                      </span>
                                      <div className="media-body">
                                          <span className="block">New feature added</span>
                                          <span className="text-muted">9 hours ago</span>
                                      </div>
                                  </a>
                              </li>
                          </ul>
                          <div className="panel-footer">
                              <a href="javascript:;">Show all notifications.</a>
                          </div>
                      </div>
                  </li>
                  <li>
                      <a href="#/tasks/tasks" className="bg-danger" id="step3">
                          <i className="fa fa-tasks"></i>
                      </a>
                  </li>
              </ul>
          </div>
      </header>
    </section>
