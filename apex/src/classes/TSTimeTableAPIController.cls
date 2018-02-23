@RestResource(urlMapping='/Dakoku/*')
global with sharing class TSTimeTableAPIController {
    public List<teamspirit__AtkEmpDay__c> empDays;
    public String empId;
    public Integer yearMonth;
    public String startDate;
    public String lastModifiedDate;
    public Integer stdStartTime;
    public Integer stdEndTime;
    public teamspirit__AtkEmpDay__c empToday;

    global class TimeTableResponse {
        public List<Map<String, Integer>> timeTable;
        public Boolean isHoliday;
    }

    @HttpGet
    global static TimeTableResponse handleGetTimeTable() {
        TSTimeTableAPIController ctrl = new TSTimeTableAPIController();
        return ctrl.getTimeTable();
    }

    @HttpPost
    global static String handleInputTimeTable(List<Map<String, Integer>> timeTable) {
        TSTimeTableAPIController ctrl = new TSTimeTableAPIController();
        if (ctrl.inputTimeTable(timeTable)) {
            return 'OK';
        }
        return 'NG';
    }

    @HttpPut
    global static String handleSetAttendance(Boolean attendance) {
        TSTimeTableAPIController ctrl = new TSTimeTableAPIController();
        if (ctrl.setAttendance(attendance)) {
            return 'OK';
        }
        return 'NG';
    }

    public TSTimeTableAPIController() {
        loadData();
    }

    public Boolean inputTimeTable(List<Map<String, Integer>> timeTable) {
        Map<String, Object> params = getBaseParams();
        params.put('timeTable', timeTable);
        String jsonReq = JSON.serialize(params);
        Map<String, Object> res = teamspirit.RtkPotalCtl.inputTimeTable(jsonReq);
        lastModifiedDate = String.valueOf(res.get('lastModifiedDate'));
        return res.get('result') == 'OK';
    }

    public Boolean setAttendance(Boolean attendance) {
        DateTime now = DateTime.now();
        Integer timeHM = now.hour() * 60 + now.minute();
        Map<String, Object> params = getBaseParams();
        Map<String, Object> input = new Map<String, Object>{'comment' => '', 'time' => timeHM, 'face' => attendance ? 0 : 1, 'fix' => false, 'type' => 10};
            params.put('input', input);
        params.put('prevFlag', false);
        params.put('stdStartTime', stdStartTime);
        params.put('stdEndTime', stdEndTime);
        String jsonReq = JSON.serialize(params);
        Map<String, Object> res = teamspirit.RtkPotalCtl.inputTime(jsonReq);
        lastModifiedDate = String.valueOf(res.get('lastModifiedDate'));
        return res.get('result') == 'OK';
    }

    public TimeTableResponse getTimeTable() {
        TimeTableResponse res = new TimeTableResponse();
        res.isHoliday = isHoliday();
        List<Map<String, Integer>> timeTable = new List<Map<String, Integer>>();
        res.timeTable = timeTable;
        if (empToday == null) {
            return res;
        }

        Map<String, Integer> item = new Map<String, Integer>{};
            item.put('from', Integer.valueOf(empToday.teamspirit__StartTime__c));
        item.put('to', Integer.valueOf(empToday.teamspirit__EndTime__c));
        item.put('type', 1);
        timeTable.add(item);

        List<String> timeTableStrItems = (empToday.teamspirit__TimeTable__c == null ? '' : empToday.teamspirit__TimeTable__c).split(':');
        for (String timeTableStr: timeTableStrItems) {
            if(timeTableStr == '') {
                continue;
            }
            item = new Map<String, Integer>{};

                String str = timeTableStr.substring(0, 4);
            str = timeTableStr.substring(0, 4);
            if (str != '----') {
                item.put('from', Integer.valueOf(str));
            }
            str = timeTableStr.substring(4, 8);
            if (str != '----') {
                item.put('to', Integer.valueOf(str));
            }
            str = timeTableStr.substring(8, 10);
            item.put('type', Integer.valueOf(str));

            timeTable.add(item);
        }

        return res;
    }

    private Date getToday() {
        return Date.today();
    }

    private void loadData() {
        Map<String, Object> lastData = teamspirit.RtkPotalCtl.getLastModifiedDate();
        empId = (String) lastData.get('empId');
        lastModifiedDate = String.valueOf(lastData.get('lastModifiedDate'));
        Map<String, Object> empMonth = teamspirit.RtkPotalCtl.loadEmpMonth('');
        List<teamspirit__AtkConfig__c > configs = (List<teamspirit__AtkConfig__c>) empMonth.get('configs');
        teamspirit__AtkConfig__c config = configs != null && configs.size() > 0 ? configs[0] : null;
        if (config != null) {
            stdStartTime = Integer.valueOf(config.teamspirit__StdStartTime__c);
            stdEndTime = Integer.valueOf(config.teamspirit__StdEndTime__c);
        }
        empDays = (List<teamspirit__AtkEmpDay__c>) empMonth.get('empDays');
        if (empDays == null) {
            empDays = new List<teamspirit__AtkEmpDay__c>();
        }
        Date today = getToday();
        for (teamspirit__AtkEmpDay__c day: empDays) {
            if (day.teamspirit__Date__c == today) {
                empToday = day;
                break;
            }
        }
        yearMonth = (Integer) empMonth.get('yearMonth');
        startDate = (String) empMonth.get('startDate');
    }

    public Boolean isHoliday() {
        return empToday != null && empToday.teamspirit__DayType__c != null && Integer.valueOf(empToday.teamspirit__DayType__c) > 0 && empToday.teamspirit__HolidayWorkApplyId__c == null;
    }

    private Map<String, Object>getBaseParams() {
        Map<String, Object> params = new Map<String, Object>();
        params.put('empId', empId);
        params.put('month', yearMonth);
        params.put('startDate', startDate);
        params.put('lastModifiedDate', lastModifiedDate);
        params.put('date', DateTime.newInstance(getToday(), Time.newInstance(0, 0, 0, 0)).format('yyyy-MM-dd'));
        return params;
    }
}
