package main

func safetyWarning() string {
	if uiArabic {
		return "تحذير: فينكورد العرب مشروع مستقل وغير تابع للفريق الرسمي.\n" +
			"مستودعات فينكورد العرب على GitHub هي المصدر الرسمي لهذه النسخة.\n" +
			"أي موقع آخر ينتحل اسم فينكورد العرب غير موثوق.\n" +
			"إذا نزّلت نسخة من مصدر غير موثوق، أزلها فورًا،\n" +
			"وافحص جهازك من البرمجيات الخبيثة وغيّر كلمة مرور Discord."
	}
	return "Warning: Vencord Arab is an independent project and is not affiliated with the upstream team.\n" +
		"Our project's GitHub repositories are the official sources for Vencord Arab.\n" +
		"Any other site impersonating Vencord Arab is untrusted.\n" +
		"If you downloaded a copy from an untrusted source, uninstall it immediately,\n" +
		"scan your device for malware and change your Discord password."
}
