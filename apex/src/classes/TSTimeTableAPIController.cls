@RestResource(urlMapping='/Dakoku/*')
global with sharing class TSTimeTableAPIController {
    public List<teamspirit__AtkEmpDay__c> empDays;
    public String empId;
    public Integer yearMonth;
    public String startDate;
    public Long lastModifiedDate;

    @HttpGet
    global static List<Map<String, Integer>> handleGetTimeTable() {
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

    public TSTimeTableAPIController() {
        loadData();
    }

    public Boolean inputTimeTable(List<Map<String, Integer>> timeTable) {
        Map<String, Object> params = getBaseParams();
        params.put('timeTable', timeTable);
        String jsonReq = JSON.serialize(params);
        Map<String, Object> res = teamspirit.RtkPotalCtl.inputTimeTable(jsonReq);
        return res.get('result') == 'OK';
    }

    public List<Map<String, Integer>>getTimeTable() {
        Date today = getToday();
        teamspirit__AtkEmpDay__c empToday = null;
        for (teamspirit__AtkEmpDay__c day: empDays) {
            if (day.teamspirit__Date__c == today) {
                empToday = day;
                break;
            }
        }

        List<Map<String, Integer>> timetable = new List<Map<String, Integer>>();
        Map<String, Integer> item = new Map<String, Integer>{};

        item.put('from', Integer.valueOf(empToday.teamspirit__StartTime__c));
        item.put('to', Integer.valueOf(empToday.teamspirit__EndTime__c));
        item.put('type', 1);
        timetable.add(item);

        List<String> timetableStrItems = empToday.teamspirit__TimeTable__c.split(':');
        for (String timetableStr: timetableStrItems) {
            if(timetableStr == '') {
                continue;
            }
            item = new Map<String, Integer>{};

            String str = timetableStr.substring(0, 4);
            str = timetableStr.substring(0, 4);
            if (str != '----') {
                item.put('from', Integer.valueOf(str));
            }
            str = timetableStr.substring(4, 8);
            if (str != '----') {
                item.put('to', Integer.valueOf(str));
            }
            str = timetableStr.substring(8, 10);
            item.put('type', Integer.valueOf(str));

            timetable.add(item);
        }

        return timetable;
    }

    private Date getToday() {
        return Date.today();
    }

    private void loadData() {
        Map<String, Object> lastData = teamspirit.RtkPotalCtl.getLastModifiedDate();
        empId = (String) lastData.get('empId');
        lastModifiedDate = (Long) lastData.get('lastModifiedDate');
        Map<String, Object> empMonth = teamspirit.RtkPotalCtl.loadEmpMonth('');
        empDays = (List<teamspirit__AtkEmpDay__c>) empMonth.get('empDays');
        yearMonth = (Integer) empMonth.get('yearMonth');
        startDate = (String) empMonth.get('startDate');
    }

    private Map<String, Object>getBaseParams() {
        Map<String, Object> params = new Map<String, Object>();
        params.put('empId', empId);
        params.put('month', yearMonth);
        params.put('startDate', startDate);
        params.put('lastModifiedDate', String.valueOf(lastModifiedDate));
        params.put('date', DateTime.newInstance(getToday(), Time.newInstance(0, 0, 0, 0)).format('yyyy-MM-dd'));
        return params;
    }
}
