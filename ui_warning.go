package main

func safetyWarning() string {
	if uiArabic {
		return "تحذير: فينكورد العرب مشروع مستقل وغير تابع للفريق الرسمي.\n" +
			"المصادر الرسمية: موقع فينكورد العرب ومستودعات المشروع على GitHub.\n" +
			"أي موقع آخر ينتحل اسم فينكورد العرب غير موثوق.\n" +
			"إذا نزّلت نسخة من مصدر غير موثوق، أزلها فورًا،\n" +
			"وافحص جهازك من البرمجيات الخبيثة وغيّر كلمة مرور Discord.\n" +
			"vencordar.com"
	}
	return "Warning: Vencord Arab is an independent project and is not affiliated with the upstream team.\n" +
		"Official sources: our website below and our project's GitHub repositories.\n" +
		"Any other site impersonating Vencord Arab is untrusted.\n" +
		"If you downloaded a copy from an untrusted source, uninstall it immediately,\n" +
		"scan your device for malware and change your Discord password.\n" +
		"vencordar.com"
}
