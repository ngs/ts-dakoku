@isTest public class TSTimeTableAPIControllerTest {
    public static testMethod void testGetTimeTable(){
        TSTimeTableAPIController ctrl = new TSTimeTableAPIController();
        ctrl.empToday = null;
        List<Map<String, Integer>> res = ctrl.getTimeTable();
        System.assert(res.size() == 0);

        ctrl.empToday = new teamspirit__AtkEmpDay__c(
            teamspirit__Date__c = Date.today(),
            teamspirit__StartTime__c = 600,
            teamspirit__EndTime__c = 1140,
            teamspirit__TimeTable__c = '0720076521:1080----21:1290132021:1440147021:'
        );
        res = ctrl.getTimeTable();
        System.assert(res.size() == 5);
    }

    public static testMethod void testIsHoliday(){
        TSTimeTableAPIController ctrl = new TSTimeTableAPIController();
        ctrl.empToday = null;
        System.assert(!ctrl.isHoliday());

        ctrl.empToday = new teamspirit__AtkEmpDay__c(
            teamspirit__Date__c = Date.today(),
            teamspirit__StartTime__c = 600,
            teamspirit__EndTime__c = 1140,
            teamspirit__TimeTable__c = '0720076521:1080----21:1290132021:1440147021:'
        );
        System.assert(!ctrl.isHoliday());

        ctrl.empToday = new teamspirit__AtkEmpDay__c(
            teamspirit__Date__c = Date.today(),
            teamspirit__DayType__c = '2',
            teamspirit__StartTime__c = 600,
            teamspirit__EndTime__c = 1140,
            teamspirit__HolidayWorkApplyId__c = 'a0BN000000XXXXXXXX',
            teamspirit__TimeTable__c = '0720076521:1080----21:1290132021:1440147021:'
        );
        System.assert(!ctrl.isHoliday());

        ctrl.empToday = new teamspirit__AtkEmpDay__c(
            teamspirit__Date__c = Date.today(),
            teamspirit__DayType__c = '2',
            teamspirit__StartTime__c = 600,
            teamspirit__EndTime__c = 1140,
            teamspirit__TimeTable__c = '0720076521:1080----21:1290132021:1440147021:'
        );
        System.assert(ctrl.isHoliday());
    }

    public static testMethod void testInputTimeTable(){
        TSTimeTableAPIController ctrl = new TSTimeTableAPIController();
        Boolean res = ctrl.inputTimeTable(new List<Map<String, Integer>>{
            new Map<String, Integer>{'from' => 600, 'to' => 1140, 'type' => 1}
        });
        System.assert(!res);
    }

    public static testMethod void testSetAttendance(){
        TSTimeTableAPIController ctrl = new TSTimeTableAPIController();
        Boolean res = ctrl.setAttendance(true);
        System.assert(!res);
    }

    public static testMethod void testHTTPVerbs(){
        String res1 = TSTimeTableAPIController.handleInputTimeTable(new List<Map<String, Integer>>{
            new Map<String, Integer>{'from' => 600, 'to' => 1140, 'type' => 1}
        });
        System.assert(res1 == 'NG');
        String res2 = TSTimeTableAPIController.handleSetAttendance(false);
        System.assert(res2 == 'NG');
        List<Map<String, Integer>> res3 = TSTimeTableAPIController.handleGetTimeTable();
        System.assert(res3.size() == 0);
    }
}
