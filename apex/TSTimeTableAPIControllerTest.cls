@isTest public class TSTimeTableAPIControllerTest {
    public static testMethod void testHandleGetTimeTable(){
        TSTimeTableAPIController ctrl = new TSTimeTableAPIController();
        teamspirit__AtkEmpDay__c obj = new teamspirit__AtkEmpDay__c(
            teamspirit__Date__c = Date.today(),
            teamspirit__StartTime__c = 600,
            teamspirit__EndTime__c = 1140,
            teamspirit__TimeTable__c = '0720076521:1080----21:1290132021:1440147021:'
        );
        ctrl.empDays = new List<teamspirit__AtkEmpDay__c>{ obj };
        List<Map<String, Integer>> res = ctrl.getTimeTable();
        System.assert(res.size() == 5);
    }

    public static testMethod void testHandleInputTimeTable(){
        TSTimeTableAPIController ctrl = new TSTimeTableAPIController();
        Boolean res = ctrl.inputTimeTable(new List<Map<String, Integer>>{
            new Map<String, Integer>{'from' => 600, 'to' => 1140, 'type' => 1}
        });
        System.assert(!res);
    }
}
