'use strict';

angular.module('app.controllers', [])

# overall control
.controller('AppCtrl', [
    '$scope', '$rootScope'
    ($scope, $rootScope) ->
        $window = $(window)

        $scope.main =
            brand: 'Transform'
            name: 'Lisa Doe' # those which uses i18n directive can not be replaced for now.

        $scope.admin =
            layout: 'wide'          # 'boxed', 'wide'
            menu: 'vertical'        # 'horizontal', 'vertical'
            fixedHeader: true       # true, false
            fixedSidebar: false     # true, false

        $scope.$watch('admin', (newVal, oldVal) ->
            # manually trigger resize event to force morris charts to resize, a significant performance impact, enable for demo purpose only
            # if newVal.menu isnt oldVal.menu || newVal.layout isnt oldVal.layout
            #      $window.trigger('resize')

            if newVal.menu is 'horizontal' && oldVal.menu is 'vertical'
                 $rootScope.$broadcast('nav:reset')
                 return
            if newVal.fixedHeader is false && newVal.fixedSidebar is true
                if oldVal.fixedHeader is false && oldVal.fixedSidebar is false
                    $scope.admin.fixedHeader = true 
                    $scope.admin.fixedSidebar = true 
                if oldVal.fixedHeader is true && oldVal.fixedSidebar is true
                    $scope.admin.fixedHeader = false 
                    $scope.admin.fixedSidebar = false 
                return
            if newVal.fixedSidebar is true
                $scope.admin.fixedHeader = true
            if newVal.fixedHeader is false 
                $scope.admin.fixedSidebar = false

            return
        , true)

        $scope.color =
            primary:    '#1BB7A0'
            success:    '#94B758'
            info:       '#56BDF1'
            infoAlt:    '#7F6EC7'
            warning:    '#F3C536'
            danger:     '#FA7B58'

])

.controller('HeaderCtrl', [
    '$scope'
    ($scope) ->

        $scope.introOptions =
            steps: [
                element: '#step1',
                intro: "<strong>Heads up!</strong> You can change the layout here"
                position: 'bottom'
            ,
                element: '#step2'
                intro: "Select a different language",
                position: 'right'
            ,
                element: '#step3'
                intro: "Runnable task App",
                position: 'left'
            ,
                element: '#step4'
                intro: "Collapsed nav for both horizontal nav and vertical nav",
                position: 'right'
            ]


])

.controller('NavContainerCtrl', [
    '$scope'
    ($scope) ->


])
.controller('NavCtrl', [
    '$scope', 'taskStorage', 'filterFilter'
    ($scope, taskStorage, filterFilter) ->
        # init
        tasks = $scope.tasks = taskStorage.get()
        $scope.taskRemainingCount = filterFilter(tasks, {completed: false}).length

        $scope.$on('taskRemaining:changed', (event, count) ->
            $scope.taskRemainingCount = count
        )
])

.controller('DashboardCtrl', [
    '$scope'
    ($scope) ->

])
